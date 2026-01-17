package auth

import (
	"cloud-sentinel-k8s/pkg/common"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// InternalClaims represents our app's JWT claims
type InternalClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
	jwt.RegisteredClaims
}

// GenerateToken creates an internal JWT for the given user
func GenerateToken(userID uint, email, name string) (string, error) {
	claims := InternalClaims{
		UserID: userID,
		Email:  email,
		Name:   name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    common.AppName,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(common.JWTSecret)
}

// ValidateToken parses and validates an internal JWT
func ValidateToken(tokenString string) (*InternalClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &InternalClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return common.JWTSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*InternalClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
