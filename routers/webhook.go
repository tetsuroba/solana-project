package routers

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"os"
	"solana/models"
	"strconv"
	"sync"
)

var logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}).WithAttrs([]slog.Attr{slog.String("service", "routers")})

var logger = slog.New(logHandler)

var transactionCache = struct {
	sync.RWMutex
	m  map[int64]models.TransactionDetails
	ID int64
}{m: make(map[int64]models.TransactionDetails)}

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

	transactionCache.Lock()
	idToUse := transactionCache.ID
	transactionCache.Unlock()
	transactionDetail, err := payload[0].GetTransactionDetails(idToUse)
	if err != nil {
		context.JSON(http.StatusOK, gin.H{"status": "ok"})
		return
	}

	transactionCache.Lock()
	transactionCache.m[transactionCache.ID] = transactionDetail
	transactionCache.ID++
	transactionCache.Unlock()

	broadcast <- transactionDetail
	context.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func ClearCacheHandler(context *gin.Context) {
	ClearCache()
	context.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func GetTransactionCacheHandler(context *gin.Context) {
	context.JSON(http.StatusOK, GetTransactionCache())
}

func GetAllTransactionsAfterIDHandler(context *gin.Context) {
	IDstr := context.Query("ID")
	ID, err := strconv.ParseInt(IDstr, 10, 64)
	if err != nil {
		logger.Error("Error parsing ID", "error", err)
		_ = context.AbortWithError(http.StatusBadRequest, err)
		return
	}
	context.JSON(http.StatusOK, GetAllTransactionsAfterSignature(ID))
}

func GetLatestIDHandler(c *gin.Context) {
	c.JSON(http.StatusOK, getLatestCacheID())
}

func GetAllTransactionsAfterSignature(ID int64) []models.TransactionDetails {
	transactionCache.RLock()
	defer transactionCache.RUnlock()
	transactions := make([]models.TransactionDetails, 0)
	for id, transaction := range transactionCache.m {
		if id > ID {
			transactions = append(transactions, transaction)
		}
	}

	return transactions
}

func ClearCache() {
	transactionCache.Lock()
	transactionCache.m = make(map[int64]models.TransactionDetails)
	transactionCache.ID = 0
	transactionCache.Unlock()
}

func GetTransactionCache() []models.TransactionDetails {
	transactionCache.RLock()
	defer transactionCache.RUnlock()
	transactions := make([]models.TransactionDetails, 0)
	for _, transaction := range transactionCache.m {
		transactions = append(transactions, transaction)
	}
	return transactions
}

func getLatestCacheID() int64 {
	transactionCache.RLock()
	defer transactionCache.RUnlock()
	return transactionCache.ID - 1
}
