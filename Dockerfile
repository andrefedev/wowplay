# Use the official Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.22 as builder

# Create and change to the application directory.
WORKDIR /app

# Retrieve application dependencies using go modules.
# Allows container builds to reuse downloaded dependencies.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Build the binary.
# -mod=readonly ensures immutable go.mod and go.sum in container builds.
RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o server ./cmd/server

# Use the official Alpine image to install certificates.
# https://hub.docker.com/_/alpine
FROM alpine:3 AS certs
RUN apk add --no-cache ca-certificates

# Build the runtime container image from scratch, copying what is needed from the two previous stages.
# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-builds
FROM scratch
# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /server
# Copy the root certificates from the certs stage
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Run the web service on container startup.
ENTRYPOINT ["/server"]
# [END run_grpc_dockerfile]
# [END cloudrun_grpc_dockerfile]