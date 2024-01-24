package services

import (
	"context"
	"fmt"
	"slices"
	"solana/clients"
	"solana/models"
	"solana/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MonitoredWalletsService struct {
	db DBService
	hc clients.HeliusClient
}

func NewMonitoredWalletsService(db DBService, hc clients.HeliusClient) *MonitoredWalletsService {
	return &MonitoredWalletsService{db: db, hc: hc}
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
	var wallets = make([]*models.MonitoredWallet, 0)

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
	webhookConfig, err := mws.hc.GetWebhookConfig()
	if err != nil {
		logger.Error("Error getting webhook config", "error", err)
		return err
	}
	webhookConfig.AccountAddresses = append(webhookConfig.AccountAddresses, wallet.PublicKey)
	webHookConfigRequest := &clients.WebhookConfigRequest{
		WebhookURL:       webhookConfig.WebhookURL,
		TransactionTypes: webhookConfig.TransactionTypes,
		AccountAddresses: webhookConfig.AccountAddresses,
		WebhookType:      webhookConfig.WebhookType,
		AuthHeader:       webhookConfig.AuthHeader,
	}

	_, err = mws.hc.UpdateWebhookConfig(webHookConfigRequest)
	if err != nil {
		logger.Error("Error updating webhook config", "error", err)
		return err
	}

	_, err = mws.db.InsertOne(context.TODO(), wallet)
	if err != nil {
		return err
	}

	return nil
}

func (mws *MonitoredWalletsService) DeleteMonitoredWallet(name string) error {
	walletConfig, err := mws.GetMonitoredWalletByName(name)
	if err != nil {
		logger.Error("Error getting wallet", "error", err)
		return err
	}

	webhookConfig, err := mws.hc.GetWebhookConfig()
	if err != nil {
		logger.Error("Error getting webhook config", "error", err)
		return err
	}

	foundIndex := utils.Find(webhookConfig.AccountAddresses, walletConfig.PublicKey)
	if foundIndex == -1 {
		logger.Error("Wallet not found in webhook config", "wallet", walletConfig.PublicKey)
		return fmt.Errorf("wallet not found in webhook config %s", walletConfig.PublicKey)
	}
	webhookConfig.AccountAddresses = slices.Delete(webhookConfig.AccountAddresses, foundIndex, foundIndex+1)
	webHookConfigRequest := &clients.WebhookConfigRequest{
		WebhookURL:       webhookConfig.WebhookURL,
		TransactionTypes: webhookConfig.TransactionTypes,
		AccountAddresses: webhookConfig.AccountAddresses,
		WebhookType:      webhookConfig.WebhookType,
		AuthHeader:       webhookConfig.AuthHeader,
	}
	_, err = mws.hc.UpdateWebhookConfig(webHookConfigRequest)
	if err != nil {
		logger.Error("Error updating webhook config", "error", err)
		return err
	}

	_, err = mws.db.DeleteOne(context.Background(), bson.D{{"name", name}})
	if err != nil {
		logger.Error("Error deleting wallet", "error", err)
		return err
	}
	return nil
}

func (mws *MonitoredWalletsService) UpdateMonitoredWallet(name string, updatedWallet *models.MonitoredWallet) (*models.MonitoredWallet, error) {
	var wallet models.MonitoredWallet

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
