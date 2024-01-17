package main

import (
	"github.com/joho/godotenv"
	"log/slog"
	"net/http"
	"os"
	"solana/handlers"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelDebug,
}))

func main() {
	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file")
		panic(err)
	}
	port := os.Getenv("PORT")
	logger.Info("Starting server on port " + port)
	http.HandleFunc("/api/webhook", handlers.WebhookHandler)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		logger.Error("Error starting server", err)
	}

}
