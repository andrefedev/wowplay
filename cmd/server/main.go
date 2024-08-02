package main

import (
	"cloud.google.com/go/storage"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"wowplay"
)

var (
	clients    *storage.Client
	signalChan = make(chan os.Signal, 1)
)

func main() {
	var err error
	clients, err = storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("InitGoogleStorage: [storage.NewClient]: %v", err)
	}

	// new router
	r := chi.NewRouter()
	// First Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(middleware.CleanPath)
	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(40 * time.Minute))

	// accept only application/json
	r.Use(middleware.AllowContentType("application/json", "multipart/form-data"))

	// Custom header's
	r.Use(middleware.SetHeader("X-Frame-Options", "deny"))
	r.Use(middleware.SetHeader("X-Content-Type-Options", "nosniff"))
	r.Use(middleware.SetHeader("Content-Type", "application/json; charset=utf-8"))

	r.Get("/ping", UploadAndDownloadTvFile())

	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// create http server
	// definimos un servidor http para poder apagarlo posteriormente.
	// chi router por defecto no tiene está función, por lo tanto, debemos
	// envolverlo con http.Server.
	// httpServer := &http.Server{Addr: ":" + port, Handler: mux}
	err = http.ListenAndServe(":"+port, r)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("ListenAndServe: %v", err)
	}
}

func UploadAndDownloadTvFile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		query := r.URL.Query()

		// * default *
		url := query.Get("url")
		from := query.Get("from")

		log.Printf("url: %v", url)
		log.Printf("from: %v", from)

		// Crear una nueva solicitud GET
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error reporting: [%v]", err)
			return
		}

		log.Printf("origin: %v", from)

		// Agregar encabezados personalizados
		req.Header.Add("Origin", from)
		req.Header.Add("Referer", from)
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

		// Realizar la solicitud
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error reporting: [%v]", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("status code error: [%v]", resp.StatusCode)
			return
		}

		// FILE NAME
		fileName, _ := wowplay.GenerateRandomString(16)
		filePath := fileName + "." + "mp4" // random.ext

		// Subir el video a Google Cloud Storage en streaming
		bucket := clients.Bucket("wowtv_videos")
		object := bucket.Object(filePath)
		writer := object.NewWriter(ctx)
		defer writer.Close()

		// Copiar el contenido directamente desde el cuerpo de la respuesta al escritor de Google Cloud Storage
		if _, err := io.Copy(writer, resp.Body); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Printf("error reporting: [%v]", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(filePath)
	}
}
