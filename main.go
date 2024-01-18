package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"solana/db"
	"solana/routers"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

func main() {
	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file")
		panic(err)
	}
	port := os.Getenv("PORT")
	salt := []byte(os.Getenv("SALT"))
	logger.Info("Starting server on port " + port)
	router := gin.Default()
	v1 := router.Group("/api")
	err = db.Init()
	if err != nil {
		logger.Error("Error initializing database", err)
		panic(err)
	}
	routers.NewWalletsRouter(db.GetDB().Database("solana").Collection("wallets"), v1, salt)
	routers.NewMonitoredWalletsRouter(db.GetDB().Database("solana").Collection("monitoredWallets"), v1)

	router.POST("/api/webhook", routers.WebhookHandler)
	err = router.Run(":" + port)
	if err != nil {
		logger.Error("Error starting server", err)
		panic(err)
	}
}
