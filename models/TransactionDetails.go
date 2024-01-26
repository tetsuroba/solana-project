package models

type TransactionDetails struct {
	Account   string `json:"account"`
	Signature string `json:"signature"`
	FromToken string `json:"fromToken"`
	ToToken   string `json:"toToken"`
	AmountIn  string `json:"amountIn"`
	AmountOut string `json:"amountOut"`
	TimeStamp int64  `json:"timeStamp"`
	Status    string `json:"status"`
	Fees      int64  `json:"fees"`
	Error     string `json:"error"`
}
