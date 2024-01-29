package models

import (
	"fmt"
	"solana/db"
	"solana/utils"
	"strconv"
	"strings"
)

type SolanaPayload struct {
	Type             string               `bson:"type"`
	Description      string               `bson:"description"`
	Events           map[string]SwapEvent `bson:"events"`
	Fee              int64                `bson:"fee"`
	FeePayer         string               `bson:"feePayer"`
	Signature        string               `bson:"signature"`
	Timestamp        int64                `bson:"timestamp"`
	TransactionError string               `bson:"transactionError"`
	AccountData      []AccountData        `bson:"accountData"`
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

type AccountData struct {
	Account             string         `json:"account"`
	NativeBalanceChange int64          `json:"nativeBalanceChange"`
	TokenBalanceChanges []TokenBalance `json:"tokenBalanceChanges"`
}

type TokenBalance struct {
	Mint           string         `json:"mint"`
	TokenAccount   string         `json:"tokenAccount"`
	UserAccount    string         `json:"userAccount"`
	RawTokenAmount RawTokenAmount `json:"rawTokenAmount"`
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

func (s *SolanaPayload) GetTransactionDetails(ID int64) (TransactionDetails, error) {
	var FromToken string
	var ToToken string
	var AmountIn string
	var AmountOut string
	var FromTokenDecimal int
	var ToTokenDecimal int
	var fromTokenSymbol string
	var toTokenSymbol string

	for _, accountData := range s.AccountData {
		if accountData.Account == s.FeePayer {
			if accountData.NativeBalanceChange < 0 {
				FromToken = utils.SOL_ADDRESS
				AmountOut = strconv.FormatInt(accountData.NativeBalanceChange*-1, 10)
				fromTokenSymbol = "SOL"
			} else if accountData.NativeBalanceChange > 0 {
				ToToken = utils.SOL_ADDRESS
				AmountIn = strconv.FormatInt(accountData.NativeBalanceChange, 10)
				toTokenSymbol = "SOL"
			}
		}
		if accountData.TokenBalanceChanges != nil && len(accountData.TokenBalanceChanges) > 0 {
			for _, tokenBalance := range accountData.TokenBalanceChanges {
				if tokenBalance.UserAccount == s.FeePayer {
					if tokenBalance.RawTokenAmount.TokenAmount[0] == '-' {
						FromToken = tokenBalance.Mint
						AmountOut = tokenBalance.RawTokenAmount.TokenAmount[1:]
						FromTokenDecimal = tokenBalance.RawTokenAmount.Decimals
					} else {
						ToToken = tokenBalance.Mint
						AmountIn = tokenBalance.RawTokenAmount.TokenAmount
						ToTokenDecimal = tokenBalance.RawTokenAmount.Decimals
					}
				}
			}
		}
	}

	if FromToken == "" {
		if len(s.Events["swap"].TokenInputs) > 0 {
			FromToken = s.Events["swap"].TokenInputs[0].Mint
			AmountIn = s.Events["swap"].TokenInputs[0].RawTokenAmount.TokenAmount
			FromTokenDecimal = s.Events["swap"].TokenInputs[0].RawTokenAmount.Decimals
		} else if len(s.Events["swap"].InnerSwaps) > 0 && len(s.Events["swap"].InnerSwaps[0].TokenInputs) > 0 {
			FromToken = s.Events["swap"].InnerSwaps[0].TokenInputs[0].Mint
			AmountIn = s.Events["swap"].InnerSwaps[0].TokenInputs[0].RawTokenAmount.TokenAmount
			FromTokenDecimal = s.Events["swap"].InnerSwaps[0].TokenInputs[0].RawTokenAmount.Decimals
		}
	}

	if ToToken == "" {
		if len(s.Events["swap"].TokenOutputs) > 0 {
			ToToken = s.Events["swap"].TokenOutputs[0].Mint
			AmountOut = s.Events["swap"].TokenOutputs[0].RawTokenAmount.TokenAmount
			ToTokenDecimal = s.Events["swap"].TokenOutputs[0].RawTokenAmount.Decimals
		} else if len(s.Events["swap"].InnerSwaps) > 0 && len(s.Events["swap"].InnerSwaps[0].TokenOutputs) > 0 {
			ToToken = s.Events["swap"].InnerSwaps[0].TokenOutputs[0].Mint
			AmountOut = s.Events["swap"].InnerSwaps[0].TokenOutputs[0].RawTokenAmount.TokenAmount
			ToTokenDecimal = s.Events["swap"].InnerSwaps[0].TokenOutputs[0].RawTokenAmount.Decimals
		}
	}

	result := db.GetDB().Database("solana").Collection("monitoredWallets").FindOne(nil, map[string]interface{}{"publicKey": s.FeePayer})
	var walletName string
	if result.Err() != nil {
		fmt.Printf("error finding wallet %s", result.Err())
		walletName = ""
	} else {
		var wallet MonitoredWallet
		err := result.Decode(&wallet)
		if err != nil {
			return TransactionDetails{}, fmt.Errorf("error decoding wallet %s", err)
		}
		walletName = wallet.Name
	}

	if ToToken == "" {
		logger.Debug("ToToken was not found", "signature", s.Signature, "description", s.Description)
		return TransactionDetails{}, fmt.Errorf("to token was not found")
	}
	if FromToken == "" {
		logger.Debug("FromToken was not found", "signature", s.Signature, "description", s.Description)
		return TransactionDetails{}, fmt.Errorf("from token was not found")
	}
	if FromToken == ToToken {
		logger.Debug("FromToken and ToToken are the same", "signature", s.Signature, "description", s.Description)
		return TransactionDetails{}, fmt.Errorf("from token and to token are the same")
	}

	if strings.Trim(s.Description, " ") != "" {
		splitString := strings.Split(s.Description, " ")
		if len(splitString) > 3 {
			if fromTokenSymbol == "" {
				fromTokenSymbol = splitString[3]
			}
		}
		if len(splitString) > 6 {
			if toTokenSymbol == "" {
				toTokenSymbol = splitString[6]
			}
		}
	}
	return TransactionDetails{
		ID:               ID,
		Account:          s.FeePayer,
		AccountName:      walletName,
		Signature:        s.Signature,
		FromToken:        FromToken,
		FromTokenSymbol:  fromTokenSymbol,
		FromTokenDecimal: FromTokenDecimal,
		ToToken:          ToToken,
		ToTokenSymbol:    toTokenSymbol,
		ToTokenDecimal:   ToTokenDecimal,
		AmountIn:         AmountIn,
		AmountOut:        AmountOut,
		TimeStamp:        s.Timestamp,
		Status:           "confirmed",
		Fees:             s.Fee,
		Error:            s.TransactionError,
		Description:      s.Description,
	}, nil
}
