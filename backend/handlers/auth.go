package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"moodtracker/db"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type authRequest struct {
	Email string `json:"email"`
}

type authResponse struct {
	Token string `json:"token"`
}

// RegisterAuthRoutes підключає маршрути /auth
func RegisterAuthRoutes(r chi.Router) {
	r.Post("/login", loginHandler)
}

// loginHandler – створює користувача (якщо нового) і повертає JWT
func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if req.Email == "" {
		http.Error(w, "email is required", http.StatusBadRequest)
		return
	}

	// Шукаємо або створюємо користувача
	var userID string
	err := db.DB.Get(&userID, "SELECT id FROM users WHERE email=$1", req.Email)
	if err == sql.ErrNoRows {
		userID = uuid.NewString()
		_, err = db.DB.Exec(
			`INSERT INTO users (id, email) VALUES ($1, $2)`,
			userID, req.Email,
		)
		if err != nil {
			http.Error(w, "failed to create user: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		http.Error(w, "db error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Створюємо JWT
	secret := []byte(os.Getenv("JWT_SECRET"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(), // термін 3 дні
	})
	tokenStr, err := token.SignedString(secret)
	if err != nil {
		http.Error(w, "failed to sign token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authResponse{Token: tokenStr})
}
