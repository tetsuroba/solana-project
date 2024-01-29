package db

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	_ "gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log/slog"
	"os"
)

type Database struct {
	*gorm.DB
}

var logHandler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}).WithAttrs([]slog.Attr{slog.String("service", "database")})

var logger = slog.New(logHandler)

var DB *mongo.Client

func Init(dbURI string) error {
	clientOptions := options.Client().ApplyURI(dbURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		logger.Error("Error connecting to MongoDB", "error", err)
		return err
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		logger.Error("Error pinging MongoDB", "error", err)
		return err
	}

	DB = client
	logger.Info("Connected to MongoDB!")
	return nil
}

func GetDB() *mongo.Client {
	return DB
}
