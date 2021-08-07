package security

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

type Claims struct {
	UserID int `json:"userID"`
	jwt.StandardClaims
}

func GenerateJWT(userID int) (string, error) {
	signingKey := []byte(os.Getenv("JWT_SECRET"))

	claims := &Claims{
		UserID:         userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 30).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}

	return ss, nil
}