package utils

import (
	"crypto/rand"
	"log"
	"math/big"
)

const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"

func RandomString(length int) string {
	b := make([]byte, length)

	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			log.Fatalf("Error during random string generation: %s", err)
		}

		b[i] = letters[num.Int64()]
	}

	return string(b)
}