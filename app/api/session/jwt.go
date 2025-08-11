package session

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

var JwtSecret = []byte("your_secret_key") // Private key to generate signature

type Claims struct {
	Username  string `json:"username"`
	UserGroup uint   `json:"usergroup"`
	IsAdmin   bool   `json:"admin"`
	jwt.StandardClaims
}

// GWT generation
func GenerateJWT(username string, userGroup uint, admin bool) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Token Time to live
	claims := &Claims{
		Username:  username,
		UserGroup: userGroup,
		IsAdmin:   admin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JwtSecret)
}

// Checking jwt
func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("token has an invalid signing method")
		}
		return JwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("token isn't valid")
	}

	// Дополнительная проверка на срок действия токена
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}
