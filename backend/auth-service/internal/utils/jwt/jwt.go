package jwt

import (
	"errors"
	"os"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func getSecret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

func GenerateAccessToken(userID, email string) (string, error) {
	claims := jwtlib.MapClaims{
		"user_id": userID,
		"email":   email,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
	}

	token := jwtlib.NewWithClaims(
		jwtlib.SigningMethodHS256,
		claims,
	)

	return token.SignedString(getSecret())
}

func GenerateRefreshToken(userID, email string) (string, error) {
	claims := jwtlib.MapClaims{
		"user_id": userID,
		"email":   email,
		"iat":     time.Now().Unix(),
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}

	token := jwtlib.NewWithClaims(
		jwtlib.SigningMethodHS256,
		claims,
	)

	return token.SignedString(getSecret())
}

func VerifyAccessToken(tokenString string) (*Claims, error) {
	token, err := jwtlib.Parse(tokenString, func(token *jwtlib.Token) (interface{}, error) {
		return getSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claimsMap, ok := token.Claims.(jwtlib.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	claims := &Claims{
		UserID:    claimsMap["user_id"].(string),
		Email:     claimsMap["email"].(string),
		IssuedAt:  int64(claimsMap["iat"].(float64)),
		ExpiresAt: int64(claimsMap["exp"].(float64)),
	}

	return claims, nil
}

func VerifyRefreshToken(tokenString string) (*Claims, error) {
	token, err := jwtlib.Parse(tokenString, func(token *jwtlib.Token) (interface{}, error) {
		return getSecret(), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claimsMap, ok := token.Claims.(jwtlib.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	claims := &Claims{
		UserID:    claimsMap["user_id"].(string),
		Email:     claimsMap["email"].(string),
		IssuedAt:  int64(claimsMap["iat"].(float64)),
		ExpiresAt: int64(claimsMap["exp"].(float64)),
	}

	return claims, nil
}