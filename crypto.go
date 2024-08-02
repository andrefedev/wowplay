package wowplay

import (
	"fmt"
	"math/big"

	"crypto/rand"
	"encoding/hex"
)

const (
	numChars = "1234567890"
)

func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("crypto.generateRandomBytes: %w", err)
	}
	return b, nil
}

func GenerateRandomString(l int) (string, error) {
	b, err := generateRandomBytes(l)
	if err != nil {
		return "", fmt.Errorf("crypto.GenerateRandomString(): [%w]", err)
	}
	return hex.EncodeToString(b), nil
}

func GenerateRandomNumberString(l int) (string, error) {
	b, err := generateRandomBytes(l)
	if err != nil {
		return "", fmt.Errorf("crypto.GenerateRandomNumberString: %w", err)
	}
	for i := range b {
		b[i] = numChars[int(b[i])%len(numChars)]
	}
	return string(b), nil
}

func GenerateRandomStockKeepingUnit() (int64, error) {
	minValue := int64(10000)
	maxValue := int64(99999)

	bg := big.NewInt(maxValue - minValue)
	value, err := rand.Int(rand.Reader, bg)
	if err != nil {
		return 0, fmt.Errorf("crypt.GenerateRandomStockKeepingUnit: [%w]", err)
	}

	return value.Int64() + minValue, nil
}
