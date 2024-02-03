package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
	"io"
	"net/http"
	"solana/clients"
	"strings"
	"sync"
	"time"
)

const (
	CREATE_POOL   = "CREATE_POOL"
	ADD_LIQUIDITY = "ADD_LIQUIDITY"
)

type WalletOccurence struct {
	Address    string   `json:"address"`
	Count      int      `json:"count"`
	Occurences []string `json:"occurences"`
}

type FirstBuyerResult struct {
	TokenAddress string   `json:"tokenAddress"`
	Addresses    []string `json:"addresses"`
}

type WalletTriangulatorService struct {
	rpc          *rpc.Client
	client       *http.Client
	heliusClient *clients.HeliusClient
}

type TransactionRecord struct {
	Type               string
	Slot               string
	BlockTimeUnix      string
	BlockTime          string
	Fee                string
	IsInner            string
	TxHash             string
	SourceOwnerAccount string
	SourceTokenAccount string
	DestOwnerAccount   string
	DestTokenAccount   string
	Amount             string
	Symbol             string
	Decimals           string
	TokenAddress       string
}

type GetBlockResult struct {
	// The blockhash of this block.
	Blockhash solana.Hash `json:"blockhash"`

	// The blockhash of this block's parent;
	// if the parent block is not available due to ledger cleanup,
	// this field will return "11111111111111111111111111111111".
	PreviousBlockhash solana.Hash `json:"previousBlockhash"`

	// The slot index of this block's parent.
	ParentSlot uint64 `json:"parentSlot"`

	// Present if "full" transaction details are requested.
	Transactions []AccountTransactionWithMeta `json:"transactions"`

	// Present if "signatures" are requested for transaction details;
	// an array of signatures, corresponding to the transaction order in the block.
	Signatures []solana.Signature `json:"signatures"`

	// Present if rewards are requested.
	Rewards []rpc.BlockReward `json:"rewards"`

	// Estimated production time, as Unix timestamp (seconds since the Unix epoch).
	// Nil if not available.
	BlockTime *solana.UnixTimeSeconds `json:"blockTime"`

	// The number of blocks beneath this block.
	BlockHeight *uint64 `json:"blockHeight"`
}

type AccountTransactionWithMeta struct {
	Meta        rpc.TransactionMeta `json:"meta"`
	Transaction AccountTransaction  `json:"transaction"`
	Version     interface{}         `json:"version"`
	Slot        int                 `json:"slot"`
	BlockTime   interface{}         `json:"blockTime"`
}

type AccountTransaction struct {
	AccountKeys []rpc.ParsedMessageAccount `json:"accountKeys"`
	Signatures  []string                   `json:"signatures"`
}

func NewWalletTriangulatorService(rpcUrl string, hc *clients.HeliusClient) *WalletTriangulatorService {
	return &WalletTriangulatorService{rpc: rpc.New(rpcUrl), client: &http.Client{Timeout: 10 * time.Second}, heliusClient: hc}
}

func (wts *WalletTriangulatorService) FindCommonAddressesInTokens(limit int, tokenAddresses []string) ([]WalletOccurence, error) {
	commonAddresses := make(map[string]WalletOccurence)

	// Create a channel to collect the results
	results := make(chan FirstBuyerResult)

	// Create a WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start a new goroutine for each tokenAddress
	for _, tokenAddress := range tokenAddresses {
		wg.Add(1)
		go func(tokenAddress string) {
			defer wg.Done()
			addresses, err := wts.GetFirstBuyersOfToken(tokenAddress, limit)
			if err != nil {
				logger.Error("Error getting first buyers of token", "error", err, "tokenAddress", tokenAddress)
				// If there's an error, send an empty slice
				addresses = []string{}
			}
			results <- FirstBuyerResult{tokenAddress, addresses}
		}(tokenAddress)
	}

	// Start a new goroutine to close the results channel after all other goroutines have finished
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect the results from the channel
	for addresses := range results {
		for _, address := range addresses.Addresses {
			if _, ok := commonAddresses[address]; !ok {
				commonAddresses[address] = WalletOccurence{Address: address, Count: 1, Occurences: []string{addresses.TokenAddress}}
			} else {
				commonAddresses[address] = WalletOccurence{Address: address, Count: commonAddresses[address].Count + 1, Occurences: append(commonAddresses[address].Occurences, addresses.TokenAddress)}
			}
		}
	}

	addressSlice := make([]WalletOccurence, 0, len(commonAddresses))
	for address := range commonAddresses {
		addressSlice = append(addressSlice, commonAddresses[address])
	}

	return addressSlice, nil
}

func (wts *WalletTriangulatorService) GetFirstBuyersOfToken(tokenAddress string, limit int) ([]string, error) {
	tokenMintTransaction, err := wts.getTokenMintTransaction(tokenAddress)
	if err != nil {
		logger.Error("Error getting token mint transaction", "error", err, "tokenAddress", tokenAddress)
		return []string{}, err
	}
	var addresses = struct {
		sync.RWMutex
		m map[string]bool
	}{m: make(map[string]bool)}

	deployerAddress := tokenMintTransaction.SourceOwnerAccount
	deploymentSignature := tokenMintTransaction.TxHash
	logger.Info("Getting first buyers of token", "tokenAddress", tokenAddress, "deployerAddress", deployerAddress, "deploymentSignature", deploymentSignature)
	deployerTransactions, err := wts.heliusClient.GetAccountTokenTransactions(deployerAddress, deploymentSignature)
	if err != nil {
		logger.Error("Error getting deployer transactions", "error", err, "tokenAddress", tokenAddress, "deployerAddress", deployerAddress, "deploymentSignature", deploymentSignature)
		return []string{}, err
	}
	var deploymentBlock uint64
	for _, transaction := range deployerTransactions {
		if transaction.TransactionType == CREATE_POOL || transaction.TransactionType == ADD_LIQUIDITY {
			deploymentBlock = transaction.Slot
		}
	}

	effortlessCalls := 0
	for len(addresses.m) < limit {
		addressLengthBefore := len(addresses.m)
		err = wts.getBlockAddresses(deploymentBlock, &addresses, tokenAddress, limit)
		addressesLengthAfter := len(addresses.m)
		if addressesLengthAfter > addressLengthBefore {
			effortlessCalls = 0
		}
		if err != nil {
			logger.Error("Error getting block addresses", "error", err, "tokenAddress", tokenAddress, "blockNumber", deploymentBlock)
			return []string{}, err
		}
		if effortlessCalls > 100 {
			logger.Error("Error getting block addresses", "error", "too many calls", "tokenAddress", tokenAddress, "blockNumber", deploymentBlock)
			break
		}
		effortlessCalls++
		deploymentBlock++
	}

	addressSlice := make([]string, 0)
	for address := range addresses.m {
		addressSlice = append(addressSlice, address)
	}
	return addressSlice, nil
}

func (wts *WalletTriangulatorService) getBlockAddresses(blockNumber uint64, addresses *struct {
	sync.RWMutex
	m map[string]bool
}, tokenAddress string, limit int) error {
	var out *GetBlockResult
	type M map[string]interface{}
	obj := M{}
	obj["encoding"] = solana.EncodingBase64
	obj["transactionDetails"] = "accounts"
	obj["rewards"] = false
	obj["maxSupportedTransactionVersion"] = 0

	params := []interface{}{blockNumber, obj}
	err := wts.rpc.RPCCallForInto(context.Background(), &out, "getBlock", params)
	if err != nil {
		if strings.Contains(err.Error(), "was skipped") {
			return nil
		}
		logger.Error("Error getting block", "error", err, "tokenAddress", tokenAddress, "blockNumber", blockNumber)
		return err
	}

	addressSlice := make([]string, 0)
	for _, transaction := range out.Transactions {
		tx := transaction.Transaction
		interactedWithToken := false
		var signerAddress string
		for _, accountKey := range tx.AccountKeys {
			if accountKey.Signer {
				signerAddress = accountKey.PublicKey.String()
			}
			if strings.ToLower(accountKey.PublicKey.String()) == strings.ToLower(tokenAddress) {
				interactedWithToken = true
				if signerAddress != "" {
					break
				}
			}
		}
		if !interactedWithToken {
			continue
		}
		addressSlice = append(addressSlice, signerAddress)
	}
	lengthBefore := len(addresses.m)
	addresses.Lock()
	for _, address := range addressSlice {
		if len(addresses.m) >= limit {
			break
		}
		addresses.m[address] = true
	}
	addresses.Unlock()
	logger.Info("Number of added addresses", "addedAddresses", len(addresses.m)-lengthBefore, "tokenAddress", tokenAddress, "blockNumber", blockNumber)
	return nil
}

func (wts *WalletTriangulatorService) getTokenMintTransaction(tokenAddress string) (*TransactionRecord, error) {
	unixSecondsNow := time.Now().Unix()
	unixSecondsForFirstQuery := 1610841600
	requestURL := fmt.Sprintf("https://api.solscan.io/v2/transfer/export_token?token_address=%s&type=mint&timefrom=%d&timeto=%d", tokenAddress, unixSecondsForFirstQuery, unixSecondsNow)

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		logger.Error("Error creating request", "error", err)
		return &TransactionRecord{}, err
	}
	resp, err := wts.client.Do(req)
	if err != nil {
		logger.Error("Error getting response", "error", err)
		return &TransactionRecord{}, err
	}
	if resp.StatusCode != http.StatusOK {
		logger.Error("Received non-200 status code", "status", resp.StatusCode)
		return &TransactionRecord{}, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logger.Error("Error closing response body", "error", err)
		}
	}(resp.Body)

	reader := csv.NewReader(resp.Body)
	reader.Comma = ','
	reader.FieldsPerRecord = 15
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	var firstTransactionRecord TransactionRecord
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.Error("Error reading csv record", "error", err)
			return &TransactionRecord{}, err
		}
		if record[0] == "type" {
			continue
		}
		if record[0] == "mint" {
			firstTransactionRecord = TransactionRecord{
				Type:               record[0],
				Slot:               record[1],
				BlockTimeUnix:      record[2],
				BlockTime:          record[3],
				Fee:                record[4],
				IsInner:            record[5],
				TxHash:             record[6],
				SourceOwnerAccount: record[7],
				SourceTokenAccount: record[8],
				DestOwnerAccount:   record[9],
				DestTokenAccount:   record[10],
				Amount:             record[11],
				Symbol:             record[12],
				Decimals:           record[13],
				TokenAddress:       record[14],
			}
			break
		}
	}
	return &firstTransactionRecord, nil
}

func (wts *WalletTriangulatorService) getTokenSignatures(tokenAddress solana.PublicKey, opts *rpc.GetSignaturesForAddressOpts) {
	transactionSignatures, err := wts.rpc.GetSignaturesForAddressWithOpts(context.Background(), tokenAddress, opts)
	if err != nil {
		logger.Error("Error getting signatures for address", "error", err)
		return
	}
	for _, signature := range transactionSignatures {
		logger.Info("Signature", "signature", signature)
	}
}
