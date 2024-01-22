package routers

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"os"
	"solana/models"
)

var logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}).WithAttrs([]slog.Attr{slog.String("service", "routers")})

var logger = slog.New(logHandler)

// WebhookHandler @Summary Webhook handler
// @Description Webhook handler
// @Tags Webhook
// @Accept json
// @Produce json
// @Success 200 {object} string
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /webhook [post]
func WebhookHandler(context *gin.Context) {
	if context.Request.Method != "POST" {
		logger.Error("Invalid request method", "method", context.Request.Method)
		_ = context.AbortWithError(http.StatusMethodNotAllowed, errors.New("invalid request method"))
		return
	}

	var payload []models.SolanaPayload
	if context.Request.Body == nil {
		logger.Error("Empty request body")
		_ = context.AbortWithError(http.StatusBadRequest, errors.New("empty request body"))
		return
	}
	err := json.NewDecoder(context.Request.Body).Decode(&payload)

	if err != nil {
		logger.Error("Error decoding request body", "error", err)
		_ = context.AbortWithError(http.StatusBadRequest, err)
		return
	}
	broadcast <- payload[0]
	context.JSON(http.StatusOK, gin.H{"status": "ok"})
}
