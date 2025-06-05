package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"moodtracker/db"
	"moodtracker/handlers"
	"moodtracker/telegram"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	err := db.NewDB()
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
		return
	}
	err = db.DB.MigrateUp()
	if err != nil {
		log.Fatalf("DB migration failed: %v", err)
		return
	}

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		// Дозволяємо доступ тільки з фронтенд-адреси (якщо потрібно, можна замінити на * для всіх)
		AllowedOrigins:   []string{"http://localhost:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // 5 хв
	}))

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Route("/api", func(r chi.Router) {
		r.Route("/auth", handlers.RegisterAuthRoutes)
		r.Route("/mood", handlers.RegisterMoodRoutes)
		r.Route("/user/telegram", handlers.RegisterTelegramRoutes)
	})

	telegram.Start(db.DB.DB)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on :%s...", port)
	http.ListenAndServe(":"+port, r)
}
