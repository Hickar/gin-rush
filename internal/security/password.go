package security

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"os"

	"golang.org/x/crypto/scrypt"
)

// https://gist.github.com/Zenithar/6f650560fe710133e24d

const (
	scryptN      = 16384
	scryptR      = 8
	scryptP      = 1
	scryptKeyLen = 32
)

func hmacSha256(in, key []byte) ([]byte, error) {
	mac := hmac.New(sha256.New, key)
	_, err := mac.Write(in)

	if err != nil {
		return nil, err
	}

	return mac.Sum(nil), nil
}

func encScrypt(in, salt []byte) ([]byte, error) {
	return scrypt.Key(in, salt, scryptN, scryptR, scryptP, scryptKeyLen)
}

func RandomBytes(count int) ([]byte, error) {
	salt := make([]byte, count)
	_, err := rand.Read(salt)
	return salt, err
}

func HashPassword(in string, salt []byte) ([]byte, error) {
	peppered, _ := hmacSha256([]byte(in), []byte(os.Getenv("HMAC_KEY")))

	cur, err := encScrypt(peppered, salt)
	if err != nil {
		return nil, err
	}

	return cur, err
}

func VerifyPassword(plain string, hashed, salt []byte) bool {
	h, _ := HashPassword(plain, salt)
	return subtle.ConstantTimeCompare(h, hashed) == 1
}