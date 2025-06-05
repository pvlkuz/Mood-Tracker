package handlers

import (
	"encoding/json"
	"net/http"

	"moodtracker/db"
	"moodtracker/middleware"

	"github.com/go-chi/chi/v5"
)

type chatReq struct {
	ChatID int64 `json:"chat_id"`
}

// RegisterTelegramRoutes реєструє POST /user/telegram
func RegisterTelegramRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuth)
		r.Post("/register", RegisterTelegram)
	})
}

func RegisterTelegram(w http.ResponseWriter, r *http.Request) {
	var req chatReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	userID := r.Context().Value(middleware.UserIDKey).(string)
	// Оновлюємо користувача
	_, err := db.DB.Exec(
		`UPDATE users SET telegram_chat_id=$1 WHERE id=$2`,
		req.ChatID, userID,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
