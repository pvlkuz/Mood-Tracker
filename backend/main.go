package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	"moodtracker/db"
	"moodtracker/handlers"
	"moodtracker/telegram"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	//db.Connect()
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
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Route("/auth", handlers.RegisterAuthRoutes)
	r.Route("/mood", handlers.RegisterMoodRoutes)
	r.Route("/user/telegram", handlers.RegisterTelegramRoutes)

	telegram.Start(db.DB.DB)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on :%s...", port)
	http.ListenAndServe(":"+port, r)
}
