package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"moodtracker/db"
	"moodtracker/middleware"
	"moodtracker/models"

	"encoding/json"

	"github.com/google/uuid"
)

func RegisterMoodRoutes(r chi.Router) {
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuth)
		r.Post("/", createMood)
		r.Get("/", listMood)
		r.Put("/{id}", updateMood)
		r.Delete("/{id}", deleteMood)
	})
}

func createMood(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Icon    string `json:"icon"`
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	now := time.Now()
	m := models.Mood{
		ID:        uuid.NewString(),
		UserID:    userID,
		Date:      now,
		Icon:      in.Icon,
		Comment:   in.Comment,
		CreatedAt: now,
		UpdatedAt: now,
	}

	query := `
        INSERT INTO mood (id, user_id, date, icon, comment, created_at, updated_at)
        VALUES (:id, :user_id, :date, :icon, :comment, :created_at, :updated_at)`
	if _, err := db.DB.NamedExec(query, &m); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

func listMood(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	baseQuery := `SELECT * FROM mood WHERE user_id=$1`
	args := []interface{}{userID}

	if from != "" && to != "" {
		baseQuery += " AND date BETWEEN $2 AND $3"
		args = append(args, from, to)
	}

	var moods []models.Mood
	if err := db.DB.Select(&moods, baseQuery, args...); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(moods)
}

func updateMood(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var in struct {
		Icon    string `json:"icon"`
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(middleware.UserIDKey).(string)
	res, err := db.DB.Exec(
		`UPDATE mood SET icon=$1, comment=$2, updated_at=$3 WHERE id=$4 AND user_id=$5`,
		in.Icon, in.Comment, time.Now(), id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func deleteMood(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := r.Context().Value(middleware.UserIDKey).(string)
	res, err := db.DB.Exec(
		`DELETE FROM mood WHERE id=$1 AND user_id=$2`,
		id, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
