package routers

import (
	"github.com/gin-gonic/gin"
	"solana/clients"
	"solana/services"
	"strconv"
)

type ScannerRouter struct {
	Helius *clients.HeliusClient
	wtr    *services.WalletTriangulatorService
}

func NewScannerRouter(rpcURL string, helius *clients.HeliusClient) *ScannerRouter {
	wtr := services.NewWalletTriangulatorService(rpcURL, helius)
	return &ScannerRouter{Helius: helius, wtr: wtr}
}

func (sr *ScannerRouter) SetupRoutes(router *gin.RouterGroup) {
	router.GET("/scanner", sr.GetFirstBuyersOfToken)
	router.GET("/scanner/commonBuyers", sr.GetCommonBuyersOfTokens)
}

func (sr *ScannerRouter) GetCommonBuyersOfTokens(c *gin.Context) {

	// Get all tokenAddress query parameters and add them to a slice
	tokenAddresses := make([]string, 0)
	i := 1
	for c.Query("tokenAddress"+strconv.Itoa(i)) != "" {
		tokenAddress := c.Query("tokenAddress" + strconv.Itoa(i))
		if tokenAddress != "" {
			tokenAddresses = append(tokenAddresses, tokenAddress)
		}
		i++
	}
	limitString := c.Query("limit")
	if limitString == "" {
		c.JSON(400, gin.H{"error": "limit is required"})
		return
	}
	limit, err := strconv.Atoi(limitString)
	if err != nil {
		c.JSON(400, gin.H{"error": "limit must be a number"})
		return
	}
	logger.Info("Getting common buyers of tokens", "tokenAddresses", tokenAddresses, "limit", limit)
	buyers, err := sr.wtr.FindCommonAddressesInTokens(limit, tokenAddresses)
	if err != nil {
		logger.Error("Error getting common buyers of tokens", "error", err, "tokenAddresses", tokenAddresses, "limit", limit)
		c.JSON(500, gin.H{"error": "Error fetching common buyers of tokens"})
		return
	}
	c.JSON(200, gin.H{"commonBuyers": buyers})
}

func (sr *ScannerRouter) GetFirstBuyersOfToken(c *gin.Context) {
	tokenAddress := c.Query("tokenAddress")
	if tokenAddress == "" {
		c.JSON(400, gin.H{"error": "tokenAddress is required"})
		return
	}
	limitString := c.Query("limit")
	if limitString == "" {
		c.JSON(400, gin.H{"error": "limit is required"})
		return
	}
	limit, err := strconv.Atoi(limitString)
	if err != nil {
		c.JSON(400, gin.H{"error": "limit must be a number"})
		return
	}
	buyers, err := sr.wtr.GetFirstBuyersOfToken(tokenAddress, limit)
	if err != nil {
		logger.Error("Error getting first buyers of token", "error", err, "tokenAddress", tokenAddress, "limit", limit)
		c.JSON(500, gin.H{"error": "Error fetching first buyers of token"})
		return
	}
	c.JSON(200, gin.H{"firstBuyers": buyers})
}
