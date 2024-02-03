package main

import (
	"github.com/gin-contrib/cors"
	"log/slog"
	"net/http"
	"os"
	"solana/clients"
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

func initDB(dbURI string) {
	err := db.Init(dbURI)
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
	rpcURL := os.Getenv("RPC_URL")

	v1 := router.Group("/api")
	auth := router.Group("/auth")
	socket := router.Group("/socket")
	transactionsCache := router.Group("/transactionCache")

	routers.SetupCachingRoutes(transactionsCache)

	v1.Use(gin.BasicAuth(basicAuthAccounts))

	socket.GET("/transactionSocket", routers.TransactionSocketHandler)
	router.POST("/login", routers.Login)
	v1.POST("/register", routers.Register)
	v1.POST("/webhook", routers.WebhookHandler)
	auth.Use(routers.AuthMiddleware())
	{
		auth.GET("", func(c *gin.Context) {
			username := c.MustGet("username").(string)
			c.JSON(http.StatusOK, gin.H{"message": "Valid token for user: " + username})
		})
	}
	routers.NewWalletsRouter(db.GetDB().Database("solana").Collection("wallets"), v1, salt)
	hc := clients.NewHeliusClient(heliusAPIKey, heliusWebhookID)
	routers.NewMonitoredWalletsRouter(db.GetDB().Database("solana").Collection("monitoredWallets"), v1, heliusAPIKey, heliusWebhookID)
	sr := routers.NewScannerRouter(rpcURL, hc)
	sr.SetupRoutes(v1)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	routers.StartWebSocketManager()

}

func main() {
	loadEnv()

	port := os.Getenv("PORT")
	logger.Info("Starting server on port " + port)
	router := setupRouter()

	dbURI := os.Getenv("MONGODB_URI")

	initDB(dbURI)

	setupRoutes(router)

	err := router.Run(":" + port)
	if err != nil {
		logger.Error("Error starting server", err)
		panic(err)
	}
}
