package models

type SolanaPayload struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Events      map[string]interface{} `json:"events"`
	Fee         int64                  `json:"fee"`
	FeePayer    string                 `json:"feePayer"`
	Signature   string                 `json:"signature"`
}
