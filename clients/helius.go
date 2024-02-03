package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"io"
	"net/http"
	"solana/models"
)

const heliusApi = "https://api.helius.xyz/v0"

type WebhookConfig struct {
	WebhookID        string   `json:"webhookID"`
	Wallet           string   `json:"wallet"`
	WebhookURL       string   `json:"webhookURL"`
	TransactionTypes []string `json:"transactionTypes"`
	AccountAddresses []string `json:"accountAddresses"`
	WebhookType      string   `json:"webhookType"`
	AuthHeader       string   `json:"authHeader"`
}

type WebhookConfigRequest struct {
	WebhookURL       string   `json:"webhookURL"`
	TransactionTypes []string `json:"transactionTypes"`
	AccountAddresses []string `json:"accountAddresses"`
	WebhookType      string   `json:"webhookType"`
	AuthHeader       string   `json:"authHeader"`
}

type HeliusClient struct {
	apiKey    string
	webhookID string
}

type HeliusTransactionResponse struct {
	Description      string                      `json:"description"`
	TransactionType  string                      `json:"type"`
	Source           string                      `json:"source"`
	Fee              uint64                      `json:"fee"`
	FeePayer         string                      `json:"feePayer"`
	Signature        string                      `json:"signature"`
	Slot             uint64                      `json:"slot"`
	Timestamp        uint64                      `json:"timestamp"`
	TokenTransfers   []models.TokenIO            `json:"tokenTransfers"`
	NativeTransfers  []models.TokenIO            `json:"nativeTransfers"`
	AccountData      []models.AccountData        `json:"accountData"`
	TransactionError interface{}                 `json:"transactionError"`
	Instructions     []solana.GenericInstruction `json:"instructions"`
	Events           interface{}                 `json:"events"`
}

func NewHeliusClient(apiKey, webhookID string) *HeliusClient {
	return &HeliusClient{apiKey: apiKey, webhookID: webhookID}
}

func (hc *HeliusClient) GetWebhookConfig() (*WebhookConfig, error) {
	logger.Info("Getting webhook config", "webhookID", hc.webhookID)
	url := fmt.Sprintf("%s/webhooks/%s?api-key=%s", heliusApi, hc.webhookID, hc.apiKey)

	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Error("Error creating webhook config request", "error", err)
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error getting webhook config", "error", err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logger.Error("Error closing response body", "error", err)
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		logger.Error("Received non-200 status code", "status", resp.StatusCode, "webhookID", hc.webhookID)
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	var config WebhookConfig
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading webhook config body", "error", err)
		return nil, err
	}

	err = json.Unmarshal(body, &config)
	if err != nil {
		logger.Error("Error unmarshalling webhook config", "error", err)
		return nil, err
	}

	return &config, nil
}

func (hc *HeliusClient) UpdateWebhookConfig(configRequest *WebhookConfigRequest) (*WebhookConfig, error) {
	logger.Info("Updating webhook config", "webhookID", hc.webhookID)
	url := fmt.Sprintf("%s/webhooks/%s?api-key=%s", heliusApi, hc.webhookID, hc.apiKey)

	requestBody, err := json.Marshal(configRequest)
	if err != nil {
		logger.Error("Error marshalling webhook config request", "error", err)
		return nil, fmt.Errorf("error marshalling webhook config request")
	}

	client := &http.Client{}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(requestBody))
	if err != nil {
		logger.Error("Error creating webhook config request", "error", err)
		return nil, fmt.Errorf("error creating webhook config request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error updating webhook config", "error", err)
		return nil, fmt.Errorf("error updating webhook config")
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logger.Error("Error closing response body", "error", err)
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error from Helius with status code: %d", resp.StatusCode)
	}

	var updatedConfig WebhookConfig
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading webhook config body", "error", err)
		return nil, fmt.Errorf("error reading webhook config body")
	}

	err = json.Unmarshal(body, &updatedConfig)
	if err != nil {
		logger.Error("Error unmarshalling webhook config", "error", err)
		return nil, fmt.Errorf("error unmarshalling webhook config")
	}

	return &updatedConfig, nil
}

func (hc *HeliusClient) GetAccountTokenTransactions(address string, mintSignature string) ([]HeliusTransactionResponse, error) {
	url := heliusApi + "/addresses/" + address + "/transactions?source=RAYDIUM&until" + mintSignature + "&api-key=" + hc.apiKey
	logger.Info("Getting account token transactions", "url", url)
	req, err := http.NewRequest("GET", url, nil)
	var transactions []HeliusTransactionResponse
	if err != nil {
		logger.Error("Error creating request", "error", err)
		return nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error getting account token transactions", "error", err)
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logger.Error("Error closing response body", "error", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		logger.Error("Received non-200 status code", "status", resp.StatusCode)
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading response body", "error", err)
		return nil, err
	}
	err = json.Unmarshal(body, &transactions)
	if err != nil {
		logger.Error("Error unmarshalling response body", "error", err)
		return nil, err
	}
	return transactions, nil
}
