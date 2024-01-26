package main

import (
	"github.com/gin-contrib/cors"
	"log/slog"
	"net/http"
	"os"
	"solana/db"
	"solana/routers"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		logger.Error("Error loading .env file")
		panic(err)
	}
}

func setupRouter() *gin.Engine {
	router := gin.Default()
	config := cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
		ExposeHeaders:    []string{"Authorization"},
	}
	router.Use(cors.New(config))
	return router
}

func initDB() {
	err := db.Init()
	if err != nil {
		logger.Error("Error initializing database", err)
		panic(err)
	}
}

func setupRoutes(router *gin.Engine) {
	basicAuthAccounts := gin.Accounts{
		os.Getenv("BASIC_AUTH_USERNAME"): os.Getenv("BASIC_AUTH_PASSWORD"),
	}
	salt := []byte(os.Getenv("SALT"))
	heliusAPIKey := os.Getenv("HELIUS_API_KEY")
	heliusWebhookID := os.Getenv("HELIUS_WEBHOOK_ID")

	v1 := router.Group("/api")
	v1.Use(gin.BasicAuth(basicAuthAccounts))
	socket := router.Group("/socket")

	router.POST("/login", routers.Login)
	auth := router.Group("/auth")
	auth.Use(routers.AuthMiddleware())
	{
		auth.GET("", func(c *gin.Context) {
			username := c.MustGet("username").(string)
			c.JSON(http.StatusOK, gin.H{"message": "Valid token for user: " + username})
		})
	}
	routers.NewWalletsRouter(db.GetDB().Database("solana").Collection("wallets"), v1, salt)
	routers.NewMonitoredWalletsRouter(db.GetDB().Database("solana").Collection("monitoredWallets"), v1, heliusAPIKey, heliusWebhookID)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	socket.GET("/transactionSocket", routers.TransactionSocketHandler)
	v1.POST("/register", routers.Register)

	routers.StartWebSocketManager()
	v1.POST("/webhook", routers.WebhookHandler)
}

func main() {
	loadEnv()

	port := os.Getenv("PORT")
	logger.Info("Starting server on port " + port)

	router := setupRouter()

	initDB()

	setupRoutes(router)

	err := router.Run(":" + port)
	if err != nil {
		logger.Error("Error starting server", err)
		panic(err)
	}
}
