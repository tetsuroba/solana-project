package routers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"solana/models"
	"solana/services"
)

func (wr *WalletsRouter) WalletRegister(router *gin.RouterGroup) {
	router.GET("/wallet/:name", wr.getWallet)
	router.POST("/wallet", wr.addWallet)
	router.DELETE("/wallet/:name", wr.deleteWallet)
	router.PUT("/wallet/:name", wr.updateWallet)
	router.GET("/wallet", wr.getAllWallets)
}

type WalletsRouter struct {
	db             *mongo.Collection
	walletsService *services.WalletsService
}

func NewWalletsRouter(db *mongo.Collection, router *gin.RouterGroup, salt []byte) *WalletsRouter {
	wr := &WalletsRouter{db: db, walletsService: services.NewWalletsService(db, salt)}
	wr.WalletRegister(router)
	return wr
}

func (wr *WalletsRouter) getWallet(c *gin.Context) {
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name parameter is required"})
		return
	}

	wallet, err := wr.walletsService.GetWalletByName(name)
	if err != nil {
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

	c.JSON(http.StatusOK, wallet)
}

func (wr *WalletsRouter) updateWallet(c *gin.Context) {
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name parameter is required"})
		return
	}

	var wallet models.Wallet

	if err := c.BindJSON(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if wallet.PublicKey == "" || wallet.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Public key and name are required"})
		return
	}

	updatedWallet, err := wr.walletsService.UpdateWallet(name, &wallet)
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

func (wr *WalletsRouter) getAllWallets(c *gin.Context) {
	wallets, err := wr.walletsService.GetAllWallets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wallets)
}

func (wr *WalletsRouter) addWallet(c *gin.Context) {
	var wallet models.Wallet

	if err := c.BindJSON(&wallet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if wallet.PublicKey == "" || wallet.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Public key and name are required"})
		return
	}

	if err := wr.walletsService.AddWallet(&wallet); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, wallet)
}

func (wr *WalletsRouter) deleteWallet(c *gin.Context) {
	name := c.Param("name")

	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name parameter is required"})
		return
	}

	err := wr.walletsService.DeleteWallet(name)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Wallet deleted successfully"})
}
