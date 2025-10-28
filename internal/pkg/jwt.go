package pkg

import (
	"errors"
	"os"
	"time"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID	uint   
	IsModerator bool   
	Login	string 
	jwt.RegisteredClaims
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func GenerateToken(userID uint, login string, isModerator bool) (string, error) {
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("fallback-secret-key")
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:	userID,
		Login:	login,
		IsModerator: isModerator,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "lab1-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenString string) (*Claims, error) {
	if len(jwtSecret) == 0 {
		jwtSecret = []byte("fallback-secret-key")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func (c *Claims) GetExpirationTime() time.Time {
	if c.ExpiresAt != nil {
		return c.ExpiresAt.Time
	}
	return time.Time{}
}