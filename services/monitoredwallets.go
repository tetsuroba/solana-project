package services

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"solana/models"
)

type MonitoredWalletsService struct {
	db DBService
}

func NewMonitoredWalletsService(db DBService) *MonitoredWalletsService {
	return &MonitoredWalletsService{db: db}
}

func (mws *MonitoredWalletsService) GetMonitoredWalletByName(name string) (*models.MonitoredWallet, error) {
	var wallet models.MonitoredWallet

	// Finding the wallet by name
	result := mws.db.FindOne(context.Background(), bson.D{{"name", name}})

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			// Handle case where no document is found
			return nil, nil
		}
		logger.Error("Error finding wallet", "error", result.Err())
		return nil, result.Err()
	}

	// Decoding the result into the wallet variable
	err := result.Decode(&wallet)
	if err != nil {
		logger.Error("Error decoding wallet", "error", err)
		return nil, err
	}

	return &wallet, nil
}

func (mws *MonitoredWalletsService) GetAllMonitoredWallets() ([]*models.MonitoredWallet, error) {
	var wallets []*models.MonitoredWallet

	// Finding all wallets
	cursor, err := mws.db.Find(context.Background(), bson.D{})
	if err != nil {
		logger.Error("Error fetching monitored wallets", "error", err)
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			logger.Error("Error closing cursor", "error", err)
			return
		}
	}(cursor, context.Background())

	// Iterating through the cursor and decoding each document
	for cursor.Next(context.Background()) {
		var wallet models.MonitoredWallet
		err := cursor.Decode(&wallet)
		if err != nil {
			logger.Error("Error decoding wallet", "error", err)
			return nil, err
		}
		wallets = append(wallets, &wallet)
	}

	// Check if the cursor encountered any errors during iteration
	if err := cursor.Err(); err != nil {
		logger.Error("Cursor iteration error", "error", err)
		return nil, err
	}

	return wallets, nil
}

func (mws *MonitoredWalletsService) AddMonitoredWallet(wallet *models.MonitoredWallet) error {

	_, err := mws.db.InsertOne(context.TODO(), wallet)
	if err != nil {
		return err
	}

	return nil
}

func (mws *MonitoredWalletsService) DeleteMonitoredWallet(name string) error {
	_, err := mws.db.DeleteOne(context.Background(), bson.D{{"name", name}})
	if err != nil {
		logger.Error("Error deleting wallet", "error", err)
		return err
	}
	return nil
}

func (mws *MonitoredWalletsService) UpdateMonitoredWallet(name string, updatedWallet *models.MonitoredWallet) (*models.MonitoredWallet, error) {
	var wallet models.MonitoredWallet

	// Finding the wallet by name
	result := mws.db.FindOne(context.Background(), bson.D{{"name", name}})

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			// Handle case where no document is found
			logger.Error("Wallet not found", "name", name)
			return nil, nil
		}
		logger.Error("Error finding wallet", "error", result.Err())
		return nil, result.Err()
	}

	// Decoding the result into the wallet variable
	err := result.Decode(&wallet)
	if err != nil {
		logger.Error("Error decoding wallet", "error", err)
		return nil, err
	}
	wallet.Name = updatedWallet.Name
	wallet.PublicKey = updatedWallet.PublicKey
	result = mws.db.FindOneAndReplace(context.Background(), bson.D{{"name", name}}, wallet)
	if result.Err() != nil {
		logger.Error("Error updating wallet", "error", result.Err())
		return nil, result.Err()
	}

	return &wallet, nil
}
