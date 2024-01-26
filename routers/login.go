package routers

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"solana/db"
	"solana/utils"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login handles user login and JWT token generation
func Login(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	result := db.GetDB().Database("solana").Collection("users").FindOne(c, bson.M{"username": creds.Username})
	if result.Err() != nil {
		logger.Error("Error finding user", "error", result.Err())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	var parsedCreds Credentials
	err := result.Decode(&parsedCreds)
	if err != nil {
		logger.Error("Error decoding user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not decode user"})
		return
	}
	if !utils.CheckPasswordHash(creds.Password, parsedCreds.Password) {
		logger.Info("Invalid password", "username", creds.Username, "password", creds.Password, "hashedPassword", parsedCreds.Password)
		logger.Error("Invalid password", "username", creds.Username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := utils.GenerateToken(creds.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func Register(c *gin.Context) {
	var creds Credentials
	if err := c.ShouldBindJSON(&creds); err != nil {
		logger.Error("Error binding JSON", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	hashedPassword, err := utils.HashPassword(creds.Password)
	if err != nil {
		logger.Error("Error hashing password", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}
	creds.Password = hashedPassword

	result, err := db.GetDB().Database("solana").Collection("users").InsertOne(c, creds)
	if err != nil {
		logger.Error("Error inserting user", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not register user"})
		return
	}
	c.JSON(http.StatusCreated, result)
}
