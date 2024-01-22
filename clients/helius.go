package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
		return nil, err
	}

	client := &http.Client{}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(requestBody))
	if err != nil {
		logger.Error("Error creating webhook config request", "error", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Error updating webhook config", "error", err)
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
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	var updatedConfig WebhookConfig
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Error reading webhook config body", "error", err)
		return nil, err
	}

	err = json.Unmarshal(body, &updatedConfig)
	if err != nil {
		logger.Error("Error unmarshalling webhook config", "error", err)
		return nil, err
	}

	return &updatedConfig, nil
}
