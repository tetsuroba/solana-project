package models

type SolanaPayload struct {
	Type        string                 `bson:"type"`
	Description string                 `bson:"description"`
	Events      map[string]interface{} `bson:"events"`
	Fee         int64                  `bson:"fee"`
	FeePayer    string                 `bson:"feePayer"`
	Signature   string                 `bson:"signature"`
}
