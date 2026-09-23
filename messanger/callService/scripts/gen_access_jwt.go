package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AccessClaims mirrors rooms access JWT.
type AccessClaims struct {
	jwt.RegisteredClaims
	UserUUID string `json:"user_id"`
}

func main() {
	secret := envOr("JWT_SECRET", "rooms-auth-secret")
	user := envOr("USER_UUID", "11111111-1111-1111-1111-111111111111")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserUUID: user,
	})
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	fmt.Print(signed)
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
