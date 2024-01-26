package models

import "errors"

type SolanaPayload struct {
	Type             string               `bson:"type"`
	Description      string               `bson:"description"`
	Events           map[string]SwapEvent `bson:"events"`
	Fee              int64                `bson:"fee"`
	FeePayer         string               `bson:"feePayer"`
	Signature        string               `bson:"signature"`
	Timestamp        int64                `bson:"timestamp"`
	TransactionError string               `bson:"transactionError"`
}

type SwapEvent struct {
	InnerSwaps   []InnerSwap   `json:"innerSwaps"`
	NativeFees   []interface{} `json:"nativeFees"`
	NativeInput  interface{}   `json:"nativeInput"`
	NativeOutput interface{}   `json:"nativeOutput"`
	TokenFees    []interface{} `json:"tokenFees"`
	TokenInputs  []TokenIO     `json:"tokenInputs"`
	TokenOutputs []TokenIO     `json:"tokenOutputs"`
}

type InnerSwap struct {
	NativeFees   []interface{} `json:"nativeFees"`
	ProgramInfo  ProgramInfo   `json:"programInfo"`
	TokenFees    []interface{} `json:"tokenFees"`
	TokenInputs  []TokenIO     `json:"tokenInputs"`
	TokenOutputs []TokenIO     `json:"tokenOutputs"`
}

type ProgramInfo struct {
	Account         string `json:"account"`
	InstructionName string `json:"instructionName"`
	ProgramName     string `json:"programName"`
	Source          string `json:"source"`
}

type TokenIO struct {
	FromTokenAccount string         `json:"fromTokenAccount"`
	FromUserAccount  string         `json:"fromUserAccount"`
	Mint             string         `json:"mint"`
	ToTokenAccount   string         `json:"toTokenAccount"`
	ToUserAccount    string         `json:"toUserAccount"`
	TokenAmount      float64        `json:"tokenAmount"`
	TokenStandard    string         `json:"tokenStandard"`
	RawTokenAmount   RawTokenAmount `json:"rawTokenAmount"`
}

type RawTokenAmount struct {
	Decimals    int    `json:"decimals"`
	TokenAmount string `json:"tokenAmount"`
}

func (s *SolanaPayload) GetTransactionDetails() (TransactionDetails, error) {
	if s == nil || s.Events == nil || len(s.Events["swap"].TokenInputs) == 0 || len(s.Events["swap"].TokenOutputs) == 0 {
		return TransactionDetails{}, errors.New("invalid SolanaPayload")
	}
	return TransactionDetails{
		Account:   s.FeePayer,
		Signature: s.Signature,
		FromToken: s.Events["swap"].TokenInputs[0].Mint,
		ToToken:   s.Events["swap"].TokenOutputs[0].Mint,
		AmountIn:  s.Events["swap"].TokenInputs[0].RawTokenAmount.TokenAmount,
		AmountOut: s.Events["swap"].TokenOutputs[0].RawTokenAmount.TokenAmount,
		TimeStamp: s.Timestamp,
		Status:    "confirmed",
		Fees:      s.Fee,
		Error:     s.TransactionError,
	}, nil
}
