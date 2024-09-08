package server

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func Auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("token")
		if err != nil {
			if err == http.ErrNoCookie {
				http.Error(w, "Отсутствует токен", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Ошибка чтения токена", http.StatusBadRequest)
			return
		}

		tknStr := c.Value
		claims := &Claims{}

		tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			if err == jwt.ErrSignatureInvalid {
				http.Error(w, "Неверная подпись токена", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Ошибка парсинга токена", http.StatusBadRequest)
			return
		}
		if !tkn.Valid {
			http.Error(w, "Недействительный токен", http.StatusUnauthorized)
			return
		}

		if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) < 30*time.Second {
			expirationTime := time.Now().Add(5 * time.Minute)
			claims.ExpiresAt = expirationTime.Unix()
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString(jwtKey)
			if err != nil {
				http.Error(w, "Ошибка обновления токена", http.StatusInternalServerError)
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:    "token",
				Value:   tokenString,
				Expires: expirationTime,
			})
		}

		next(w, r)
	})
}
