package main

import (
	"log/slog"
	"os"
	"solana/db"
	"solana/routers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
	basicAuthAccounts := gin.Accounts{
		os.Getenv("BASIC_AUTH_USERNAME"): os.Getenv("BASIC_AUTH_PASSWORD"),
	}
	port := os.Getenv("PORT")
	salt := []byte(os.Getenv("SALT"))
	heliusAPIKey := os.Getenv("HELIUS_API_KEY")
	heliusWebhookID := os.Getenv("HELIUS_WEBHOOK_ID")
	logger.Info("Starting server on port " + port)
	router := gin.Default()
	router.Use(cors.Default())
	router.Use(gin.BasicAuth(basicAuthAccounts))
	v1 := router.Group("/api")
	err = db.Init()
	if err != nil {
		logger.Error("Error initializing database", err)
		panic(err)
	}
	routers.NewWalletsRouter(db.GetDB().Database("solana").Collection("wallets"), v1, salt)
	routers.NewMonitoredWalletsRouter(db.GetDB().Database("solana").Collection("monitoredWallets"), v1, heliusAPIKey, heliusWebhookID)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	v1.GET("/transactionSocket", routers.TransactionSocketHandler)
	routers.StartWebSocketManager()

	v1.GET("/auth", func(context *gin.Context) {
		context.JSON(200, gin.H{"status": "ok"})
	})
	v1.POST("/webhook", routers.WebhookHandler)
	err = router.Run(":" + port)
	if err != nil {
		logger.Error("Error starting server", err)
		panic(err)
	}
}
