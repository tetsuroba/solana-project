package routers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"solana/clients"
	"solana/models"
	"solana/services"
)

func (mwr *MonitoredWalletsRouter) MonitoredWalletRegister(router *gin.RouterGroup) {
	router.GET("/monitored_wallets/:name", mwr.getMonitoredWallet)
	router.POST("/monitored_wallets", mwr.addMonitoredWallet)
	router.DELETE("/monitored_wallets/:name", mwr.deleteMonitoredWallet)
	router.PUT("/monitored_wallets/:name", mwr.updateMonitoredWallet)
	router.GET("/monitored_wallets", mwr.getAllMonitoredWallets)
}

type MonitoredWalletsRouter struct {
	monitoredWalletsService *services.MonitoredWalletsService
}

func NewMonitoredWalletsRouter(db *mongo.Collection, router *gin.RouterGroup, heliusApiKey, heliusWebhookID string) *MonitoredWalletsRouter {
	hc := clients.NewHeliusClient(heliusApiKey, heliusWebhookID)
	mwr := &MonitoredWalletsRouter{monitoredWalletsService: services.NewMonitoredWalletsService(db, *hc)}
	mwr.MonitoredWalletRegister(router)
	return mwr
}

// getMonitoredWallet @Summary Get a monitored wallet by name
// @Description Get a monitored wallet by name
// @Tags Monitored Wallets
// @Param name path string true "Monitored wallet name"
// @Success 200 {object} Wallet
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /monitored_wallets/{name} [get]
func (mwr *MonitoredWalletsRouter) getMonitoredWallet(c *gin.Context) {
	// Extracting the name from the URL parameter
	name := c.Param("name")

	// Handling empty name case
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name parameter is required"})
		return
	}

	// Calling the service function
	wallet, err := mwr.monitoredWalletsService.GetMonitoredWalletByName(name)
	if err != nil {
		// Handle specific errors like not found, etc.
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if wallet == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	// Returning the found wallet
	c.JSON(http.StatusOK, wallet)
}

// updateMonitoredWallet @Summary Update a monitored wallet
// @Description Update a monitored wallet
// @Tags Monitored Wallets
// @Param name path string true "Monitored wallet name"
// @Param wallet body Wallet true "Monitored wallet object"
// @Success 200 {object} Wallet
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /monitored_wallets/{name} [put]
func (mwr *MonitoredWalletsRouter) updateMonitoredWallet(c *gin.Context) {
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name parameter is required"})
		return
	}

	var wallet models.MonitoredWallet

	// Unmarshal the JSON body into the MonitoredWallet struct
	if err := c.BindJSON(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Validate the MonitoredWallet data
	if wallet.PublicKey == "" || wallet.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Public key and name are required"})
		return
	}

	updatedWallet, err := mwr.monitoredWalletsService.UpdateMonitoredWallet(name, &wallet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	if updatedWallet == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		return
	}

	c.JSON(http.StatusOK, updatedWallet)
}

// getAllMonitoredWallets @Summary Get all monitored wallets
// @Description Get all monitored wallets
// @Tags Monitored Wallets
// @Success 200 {array} Wallet
// @Failure 500 {object} Error
func (mwr *MonitoredWalletsRouter) getAllMonitoredWallets(c *gin.Context) {
	// Calling the service function
	wallets, err := mwr.monitoredWalletsService.GetAllMonitoredWallets()
	if err != nil {
		// Handle specific errors like not found, etc.
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Returning the found wallet
	c.JSON(http.StatusOK, wallets)
}

// addMonitoredWallet @Summary Add a monitored wallet
// @Description Add a monitored wallet
// @Tags Monitored Wallets
// @Param wallet body Wallet true "Monitored wallet object"
// @Success 201 {object} Wallet
// @Failure 400 {object} Error
// @Failure 500 {object} Error
// @Router /monitored_wallets [post]
func (mwr *MonitoredWalletsRouter) addMonitoredWallet(c *gin.Context) {
	var wallet models.MonitoredWallet

	// Unmarshal the JSON body into the MonitoredWallet struct
	if err := c.BindJSON(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Validate the MonitoredWallet data
	if wallet.PublicKey == "" || wallet.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Public key and name are required"})
		return
	}

	// Call the AddMonitoredWallet service function
	if err := mwr.monitoredWalletsService.AddMonitoredWallet(&wallet); err != nil {
		// Handle different types of errors accordingly
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, wallet)
}

// deleteMonitoredWallet @Summary Delete a monitored wallet
// @Description Delete a monitored wallet
// @Tags Monitored Wallets
// @Param name path string true "Monitored wallet name"
// @Success 200 {object} Message
// @Failure 400 {object} Error
// @Failure 404 {object} Error
// @Failure 500 {object} Error
// @Router /monitored_wallets/{name} [delete]
func (mwr *MonitoredWalletsRouter) deleteMonitoredWallet(c *gin.Context) {
	// Extracting the name from the URL parameter
	name := c.Param("name")

	// Handling empty name case
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name parameter is required"})
		return
	}

	// Calling the service function
	err := mwr.monitoredWalletsService.DeleteMonitoredWallet(name)
	if err != nil {
		// Handle specific errors like not found, etc.
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Wallet deleted successfully"})
}
