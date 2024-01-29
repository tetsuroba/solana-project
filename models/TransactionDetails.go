package models

type TransactionDetails struct {
	ID               int64  `json:"id"`
	Account          string `json:"account"`
	AccountName      string `json:"accountName"`
	Signature        string `json:"signature"`
	FromToken        string `json:"fromToken"`
	FromTokenSymbol  string `json:"fromTokenSymbol"`
	FromTokenDecimal int    `json:"fromTokenDecimal"`
	ToToken          string `json:"toToken"`
	ToTokenSymbol    string `json:"toTokenSymbol"`
	ToTokenDecimal   int    `json:"toTokenDecimal"`
	AmountIn         string `json:"amountIn"`
	AmountOut        string `json:"amountOut"`
	TimeStamp        int64  `json:"timeStamp"`
	Status           string `json:"status"`
	Fees             int64  `json:"fees"`
	Error            string `json:"error"`
	Description      string `json:"description"`
}
