package jwt

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

const refreshTokenTTL = 7 * 24 * time.Hour

// Claims holds validated JWT payload fields.
type Claims struct {
	UserID    string
	Email     string
	TokenID   string
	IssuedAt  int64
	ExpiresAt int64
}

func secret() []byte {
	return []byte(os.Getenv("JWT_SECRET"))
}

// GenerateAccessToken issues a short-lived access token.
func GenerateAccessToken(userID, email string) (string, error) {
	now := time.Now()
	claims := jwtlib.MapClaims{
		"user_id": userID,
		"email":   email,
		"iat":     now.Unix(),
		"exp":     now.Add(15 * time.Minute).Unix(),
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	return token.SignedString(secret())
}

// GenerateRefreshToken issues a refresh token with a unique token id for session rotation.
func GenerateRefreshToken(userID, email string) (string, string, error) {
	tokenID := uuid.NewString()
	now := time.Now()
	claims := jwtlib.MapClaims{
		"user_id": userID,
		"email":   email,
		"jti":     tokenID,
		"iat":     now.Unix(),
		"exp":     now.Add(refreshTokenTTL).Unix(),
	}

	token := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := token.SignedString(secret())
	if err != nil {
		return "", "", err
	}

	return signed, tokenID, nil
}

// RefreshTokenTTL returns the refresh token lifetime for Redis session TTL.
func RefreshTokenTTL() time.Duration {
	return refreshTokenTTL
}

func parseClaims(tokenString string) (*Claims, error) {
	token, err := jwtlib.Parse(tokenString, func(token *jwtlib.Token) (interface{}, error) {
		if token.Method != jwtlib.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret(), nil
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

	userID, err := claimString(claimsMap, "user_id")
	if err != nil {
		return nil, err
	}
	email, err := claimString(claimsMap, "email")
	if err != nil {
		return nil, err
	}

	claims := &Claims{
		UserID:    userID,
		Email:     email,
		TokenID:   claimStringOptional(claimsMap, "jti"),
		IssuedAt:  claimInt64(claimsMap, "iat"),
		ExpiresAt: claimInt64(claimsMap, "exp"),
	}

	return claims, nil
}

// VerifyAccessToken validates an access token and returns claims.
func VerifyAccessToken(tokenString string) (*Claims, error) {
	return parseClaims(tokenString)
}

// VerifyRefreshToken validates a refresh token and returns claims including jti.
func VerifyRefreshToken(tokenString string) (*Claims, error) {
	claims, err := parseClaims(tokenString)
	if err != nil {
		return nil, err
	}
	if claims.TokenID == "" {
		return nil, errors.New("missing token id")
	}
	return claims, nil
}

func claimString(claims jwtlib.MapClaims, key string) (string, error) {
	value, ok := claims[key]
	if !ok {
		return "", fmt.Errorf("missing claim %s", key)
	}
	str, ok := value.(string)
	if !ok || str == "" {
		return "", fmt.Errorf("invalid claim %s", key)
	}
	return str, nil
}

func claimStringOptional(claims jwtlib.MapClaims, key string) string {
	value, ok := claims[key]
	if !ok {
		return ""
	}
	str, ok := value.(string)
	if !ok {
		return ""
	}
	return str
}

func claimInt64(claims jwtlib.MapClaims, key string) int64 {
	value, ok := claims[key]
	if !ok {
		return 0
	}
	switch v := value.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}
