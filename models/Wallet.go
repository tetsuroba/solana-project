package models

type Wallet struct {
	PrivateKey string `bson:"-"`
	PublicKey  string `bson:"publicKey"`
	Name       string `bson:"name"`
}
