package services

import (
	"context"
	"solana/models"
	"solana/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type WalletsService struct {
	db   DBService
	salt []byte
}

func NewWalletsService(db DBService, salt []byte) *WalletsService {
	return &WalletsService{db: db, salt: salt}
}

func (ws *WalletsService) GetWalletByName(name string) (*models.Wallet, error) {
	var wallet models.Wallet

	result := ws.db.FindOne(context.Background(), bson.D{{"name", name}})

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			// Handle case where no document is found
			return nil, nil
		}
		logger.Error("Error finding wallet", "error", result.Err())
		return nil, result.Err()
	}

	err := result.Decode(&wallet)
	if err != nil {
		logger.Error("Error decoding wallet", "error", err)
		return nil, err
	}

	return &wallet, nil
}

func (ws *WalletsService) GetAllWallets() ([]*models.Wallet, error) {
	var wallets = make([]*models.Wallet, 0)

	cursor, err := ws.db.Find(context.Background(), bson.D{})
	if err != nil {
		logger.Error("Error fetching wallets", "error", err)
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
			logger.Error("Error closing cursor", "error", err)
			return
		}
	}(cursor, context.Background())

	for cursor.Next(context.Background()) {
		var wallet models.Wallet
		err := cursor.Decode(&wallet)
		if err != nil {
			logger.Error("Error decoding wallet", "error", err)
			return nil, err
		}
		wallets = append(wallets, &wallet)
	}

	if err := cursor.Err(); err != nil {
		logger.Error("Cursor iteration error", "error", err)
		return nil, err
	}

	logger.Info("Found number of wallets", "count", len(wallets))
	return wallets, nil
}

func (ws *WalletsService) AddWallet(wallet *models.Wallet) error {
	hashedPrivateKey, err := utils.HashString(ws.salt, wallet.PrivateKey)
	if err != nil {
		logger.Error("Error hashing private key", "error", err)
		return err
	}
	wallet.PrivateKey = hashedPrivateKey
	_, err = ws.db.InsertOne(context.TODO(), wallet)
	if err != nil {
		return err
	}

	return nil
}

func (ws *WalletsService) DeleteWallet(name string) error {
	_, err := ws.db.DeleteOne(context.Background(), bson.D{{"name", name}})
	if err != nil {
		logger.Error("Error deleting wallet", "error", err)
		return err
	}
	return nil
}

func (ws *WalletsService) UpdateWallet(name string, updatedWallet *models.Wallet) (*models.Wallet, error) {
	var wallet models.Wallet

	result := ws.db.FindOne(context.Background(), bson.D{{"name", name}})

	if result.Err() != nil {
		if result.Err() == mongo.ErrNoDocuments {
			logger.Error("Wallet not found", "name", name)
			return nil, nil
		}
		logger.Error("Error finding wallet", "error", result.Err())
		return nil, result.Err()
	}

	err := result.Decode(&wallet)
	if err != nil {
		logger.Error("Error decoding wallet", "error", err)
		return nil, err
	}
	wallet.Name = updatedWallet.Name
	wallet.PublicKey = updatedWallet.PublicKey
	hashedPrivateKey, err := utils.HashString(ws.salt, updatedWallet.PrivateKey)
	if err != nil {
		logger.Error("Error hashing private key", "error", err)
		return nil, err
	}
	wallet.PrivateKey = hashedPrivateKey
	result = ws.db.FindOneAndReplace(context.Background(), bson.D{{"name", name}}, wallet)
	if result.Err() != nil {
		logger.Error("Error updating wallet", "error", result.Err())
		return nil, result.Err()
	}

	return &wallet, nil
}
