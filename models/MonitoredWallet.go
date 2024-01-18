package models

type MonitoredWallet struct {
	PublicKey string `bson:"publicKey"`
	Name      string `bson:"name"`
}
