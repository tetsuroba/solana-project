package models

import (
	"encoding/json"
	"log"
	"os"
	"testing"
)

func TestSolanaPayloadUnmarshalling(t *testing.T) {
	body, err := os.ReadFile("../test_data/solana_swap_example.json")
	if err != nil {
		log.Fatalf("unable to read file: %v", err)
	}

	var payload []SolanaPayload
	err = json.Unmarshal(body, &payload)
	if err != nil {
		t.Errorf("Unmarshalling failed: %v", err)
	}

	// Test for presence and correctness of fields
	if payload[0].Type != "SWAP" {
		t.Errorf("Expected transaction type 'SWAP', got '%s'", payload[0].Type)
	}
	if payload[0].Description == "" {
		t.Error("Expected non-empty description")
	}
}
