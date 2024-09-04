package server

import (
	"encoding/json"
	"net/http"
	"time"

	"final_sprint/config"
	"github.com/dgrijalva/jwt-go"
)

var jwtKey = []byte("your_secret_key")

type Credentials struct {
	Password string `json:"password"`
}

type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Ошибка декодирования JSON", http.StatusBadRequest)
		return
	}

	expectedPassword := config.Config["TODO_PASSWORD"]
	if creds.Password != expectedPassword {
		http.Error(w, "Неверный пароль", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		Username: "user",
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Ошибка создания токена", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
	})

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})

}
