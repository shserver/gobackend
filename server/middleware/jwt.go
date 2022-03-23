package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	secretKey            = ""
	accessTokenDuration  = 30 * time.Minute
	refreshTokenDuration = 30 * 24 * time.Hour
	refreshTokenReissue  = 10 * 24 * time.Hour
)

type MyCustomClaims struct {
	jwt.RegisteredClaims
}

// key : secret key for jwt
func LoadSecretKey(key string) {
	secretKey = key
}

// token : "access" or "refresh"
func TokenDuration(token string) time.Duration {
	if token == "access" {
		return accessTokenDuration
	} else {
		return refreshTokenDuration
	}
}

// Refresh token will be reissued when 10days is left until expiration date.
func RefreshTokenReissue(expirationDate time.Time) bool {
	if refreshTokenReissue > expirationDate.Sub(time.Now()) {
		return true
	}
	return false
}

// user: user id, duration: use TokenDuration
func CreateJWT(user string, duration time.Duration) (string, error) {
	claims := &MyCustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "sehyoung",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			Audience:  []string{user},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

// bearerToken: "Bearer ..." => verify token whether to be valid or not.
func VerifyJWT(bearerToken string) (*MyCustomClaims, error) {
	splitToken := strings.Split(bearerToken, "Bearer ")
	if len(splitToken) != 2 {
		return nil, fmt.Errorf("Bearer token Incorrect format")
	}

	token, err := jwt.ParseWithClaims(splitToken[1], &MyCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if claims, ok := token.Claims.(*MyCustomClaims); ok && token.Valid {
		return claims, nil
	} else if err == jwt.ErrTokenMalformed {
		return nil, fmt.Errorf("That's not even a token: %w", err)
	} else if err == jwt.ErrTokenExpired {
		return nil, fmt.Errorf("Token is expired: %w", err)
	} else if err == jwt.ErrTokenNotValidYet {
		return nil, fmt.Errorf("Token is not active yet: %w", err)
	} else {
		return nil, fmt.Errorf("Couldn't handle this token: %w", err)
	}
}
