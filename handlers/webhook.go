package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"solana/models"
)

var logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}).WithAttrs([]slog.Attr{slog.String("service", "webhook-handler")})

var logger = slog.New(logHandler)

func WebhookHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		logger.Error("Invalid request method", "method", r.Method)
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var payload []models.SolanaPayload
	err := json.NewDecoder(r.Body).Decode(&payload)

	if err != nil {
		logger.Error("Error decoding request body", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logger.Info("Received transaction with signature ", "signature", payload[0].Signature)
}
