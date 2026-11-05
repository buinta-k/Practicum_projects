package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var secret = []byte("my_secret_key")

func JWT(pass string) (string, error) {
	result := sha256.Sum256([]byte(pass))
	hashString := hex.EncodeToString(result[:])

	claims := jwt.MapClaims{
		"pass": hashString,
	}

	jwtToken := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	signedToken, err := jwtToken.SignedString(secret)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ValidateJWT(tokenString string) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("Не тот метод")
		}

		return secret, nil
	})

	if err != nil || !token.Valid {
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	tokenPasswordHash, ok := claims["pass"].(string)
	if !ok {
		return false
	}

	currentPassword := os.Getenv("TODO_PASSWORD")

	result := sha256.Sum256([]byte(currentPassword))
	currentPasswordHash := hex.EncodeToString(result[:])

	return tokenPasswordHash == currentPasswordHash
}

func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		pass := os.Getenv("TODO_PASSWORD")

		if len(pass) == 0 {
			next(res, req)
			return
		}

		cookie, err := req.Cookie("token")
		if err != nil {
			http.Error(
				res,
				"Authentification required",
				http.StatusUnauthorized,
			)
			return
		}

		if !ValidateJWT(cookie.Value) {
			http.Error(
				res,
				"Authentification required",
				http.StatusUnauthorized,
			)
			return
		}

		next(res, req)
	})
}

func (s *Api) SignHandler(res http.ResponseWriter, req *http.Request) {
	var pass password

	err := json.NewDecoder(req.Body).Decode(&pass)
	if err != nil {
		writeError(res, "Ошибка парсинга пароля", http.StatusBadRequest)
		return
	}

	password := os.Getenv("TODO_PASSWORD")

	if pass.Password != password {
		writeError(res, "Неверный пароль", http.StatusUnauthorized)
		return
	}

	token, err := JWT(pass.Password)
	if err != nil {
		writeError(res, "Ошибка создания токена", http.StatusInternalServerError)
		return
	}

	writeJson(res, map[string]string{
		"token": token,
	})
}
