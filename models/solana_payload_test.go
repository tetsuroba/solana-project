package models

import (
	"encoding/json"
	"testing"
)

func TestSolanaPayloadUnmarshalling(t *testing.T) {
	jsonInput := `{
        "accountData": [],
        "description": "5nM1CTQwKXFZo5yJYC8J1pgj32JW6Fx8DpQAtPZ8aiLw swapped 9.11111 USDC for 9.121867 USDC",
        "events": {"swap": {}},
        "fee": 23219,
        "feePayer": "5nM1CTQwKXFZo5yJYC8J1pgj32JW6Fx8DpQAtPZ8aiLw",
        "instructions": [],
        "nativeTransfers": [],
        "signature": "sfW8CComDyJH82uPNUZCxrtPdL8SkpVDQBQQmGSX97bgWm1MY3X6NYmkTKbZHRAb8C5dPvSknQjfQKRmL2vxh1d",
        "slot": 242424604,
        "source": "JUPITER",
        "timestamp": 1705522253,
        "tokenTransfers": [],
        "transactionError": null,
        "type": "SWAP"
    }`

	var payload SolanaPayload
	err := json.Unmarshal([]byte(jsonInput), &payload)
	if err != nil {
		t.Errorf("Unmarshalling failed: %v", err)
	}

	// Test for presence and correctness of fields
	if payload.Type != "SWAP" {
		t.Errorf("Expected transaction type 'SWAP', got '%s'", payload.Type)
	}
	if payload.Description == "" {
		t.Error("Expected non-empty description")
	}
}
