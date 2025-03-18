package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
)

const (
	rpcHost = "localhost:48332"
	rpcUser = "admin"
	rpcPass = "admin"
)

var watchedAddresses = []string{
	"muCmmr3fwCvbFbdPUgtw6KFyx92qtDyuyx",
	"mt7Wd4k9KSs6f7XtAZY96JTsPfxmZLWNMN",
	"mydBSdJF1fDfe34VJJ5v65cAtrm8w6QBW9",
}

// UTXO represents an unspent output.
type UTXO struct {
	TxID         string  `json:"txid"`
	Vout         int     `json:"vout"`
	Amount       float64 `json:"amount"`
	ScriptPubKey string  `json:"scriptPubKey"`
	Height       int     `json:"height"` // Block height where UTXO was confirmed.
	Desc         string  `json:"desc"`
}

// Transaction holds basic transaction details.
type Transaction struct {
	TxID      string   `json:"txid"`
	BlockHash string   `json:"blockhash"`
	Inputs    []string `json:"inputs"`
	Outputs   []UTXO   `json:"outputs"`
}

// Cache stores UTXOs and transactions in memory.
type Cache struct {
	utxoMap     map[string][]UTXO // Keyed by address.
	txMap       map[string]Transaction
	blockHeight int64
	mu          sync.RWMutex
}

func newCache() *Cache {
	return &Cache{
		utxoMap: make(map[string][]UTXO),
		txMap:   make(map[string]Transaction),
	}
}

func (c *Cache) updateBlockHeight(height int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if height > c.blockHeight {
		c.blockHeight = height
	}
}

func (c *Cache) getBlockHeight() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.blockHeight
}

func (c *Cache) addUTXO(address string, utxo UTXO) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.utxoMap[address] = append(c.utxoMap[address], utxo)
}

func (c *Cache) addTransaction(tx Transaction) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.txMap[tx.TxID] = tx
}

func (c *Cache) listUTXOs() []UTXO {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var utxos []UTXO
	for _, addrUTXOs := range c.utxoMap {
		utxos = append(utxos, addrUTXOs...)
	}
	return utxos
}

func (c *Cache) listTransactions() []Transaction {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var txs []Transaction
	for _, tx := range c.txMap {
		txs = append(txs, tx)
	}
	return txs
}

func main() {
	client, err := connectRPC()
	if err != nil {
		fmt.Printf("Failed to connect to Bitcoin RPC: %v\n", err)
		os.Exit(1)
	}
	defer client.Shutdown()

	fmt.Println("Starting Bitcoin UTXO Indexer...")
	cache := newCache()

	// Run a full UTXO scan.
	if err := startFullScan(client, cache); err != nil {
		fmt.Printf("Failed to perform full UTXO scan: %v\n", err)
		os.Exit(1)
	}

	// Start background monitoring for new blocks.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go continuousUTXOMonitor(ctx, client, cache)

	// On exit, print the cached UTXOs and transactions.
	defer func() {
		fmt.Println("UTXOs:")
		for _, u := range cache.listUTXOs() {
			fmt.Printf("TxID: %s, Amount: %.8f\n", u.TxID, u.Amount)
		}
		fmt.Println("Transactions:")
		for _, t := range cache.listTransactions() {
			fmt.Printf("TxID: %s, BlockHash: %s\n", t.TxID, t.BlockHash)
		}
	}()

	// Block forever.
	select {}
}

// connectRPC connects to the bitcoind RPC.
func connectRPC() (*rpcclient.Client, error) {
	connCfg := &rpcclient.ConnConfig{
		Host:         rpcHost,
		User:         rpcUser,
		Pass:         rpcPass,
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	return rpcclient.New(connCfg, nil)
}

// startFullScan performs a full UTXO scan using scantxoutset.
func startFullScan(client *rpcclient.Client, cache *Cache) error {
	fmt.Println("Performing full blockchain UTXO scan...")

	var scanObjects []map[string]interface{}
	for _, addr := range watchedAddresses {
		scanObjects = append(scanObjects, map[string]interface{}{
			"desc": fmt.Sprintf("addr(%s)", addr),
		})
	}

	rawResult, err := client.RawRequest("scantxoutset", []json.RawMessage{
		json.RawMessage(`"start"`),
		json.RawMessage(marshalJSON(scanObjects)),
	})
	if err != nil {
		return fmt.Errorf("scantxoutset failed: %v", err)
	}

	var result struct {
		Unspents  []UTXO  `json:"unspents"`
		Success   bool    `json:"success"`
		TXOuts    int     `json:"txouts"`
		Height    int64   `json:"height"`
		BestBlock string  `json:"bestblock"`
		TotalAmt  float64 `json:"total_amount"`
	}
	if err := json.Unmarshal(rawResult, &result); err != nil {
		return fmt.Errorf("failed to parse scantxoutset result: %v", err)
	}

	for _, utxo := range result.Unspents {
		scriptBytes, err := hex.DecodeString(utxo.ScriptPubKey)
		if err != nil {
			fmt.Printf("Warning: could not decode script %s: %v\n", utxo.ScriptPubKey, err)
			continue
		}
		addr, err := extractAddress(scriptBytes)
		if err != nil || addr == "" {
			fmt.Printf("Warning: could not extract address from script %s: %v\n", utxo.ScriptPubKey, err)
			continue
		}

		fmt.Printf("Found UTXO for address %s, Amount: %f\n", addr, utxo.Amount)
		cache.addUTXO(addr, utxo)

		// Retrieve transaction details for the UTXO.
		tx, txHeight, err := getTransactionDetails(client, utxo.TxID, utxo.Vout, utxo.Height)
		if err != nil {
			fmt.Printf("Warning: could not get transaction details for %s: %v\n", utxo.TxID, err)
			continue
		}
		cache.updateBlockHeight(txHeight)
		cache.addTransaction(tx)
	}

	fmt.Printf("Indexed %d UTXOs from full scan\n", len(result.Unspents))
	return nil
}

// getTransactionDetails tries to retrieve transaction details; if the direct call fails,
// it uses the UTXO confirmation height to fetch the block and extract the transaction.
func getTransactionDetails(client *rpcclient.Client, txid string, vout, height int) (Transaction, int64, error) {
	txHash, err := chainhash.NewHashFromStr(txid)
	if err != nil {
		return Transaction{}, 0, fmt.Errorf("invalid txid: %v", err)
	}

	// Try direct retrieval (mempool or wallet tx).
	txVerbose, err := client.GetRawTransactionVerbose(txHash)
	if err == nil {
		return extractTransactionDetails(client, txVerbose)
	}

	// Retrieve block hash using the UTXO's confirmation height.
	blockHash, _, err := getBlockHashForUTXO(client, txid, vout, height)
	if err != nil {
		return Transaction{}, 0, fmt.Errorf("could not get block hash for UTXO %s:%d: %v", txid, vout, err)
	}

	// Get the transaction from the block.
	tx, _, err := getTransactionFromBlock(client, txid, blockHash)
	if err != nil {
		return Transaction{}, 0, fmt.Errorf("failed to retrieve transaction %s from block %s: %v", txid, blockHash, err)
	}

	return extractTransactionDetails(client, &tx)
}

// getBlockHashForUTXO uses the confirmation height to get the block hash.
func getBlockHashForUTXO(client *rpcclient.Client, txid string, vout, height int) (*chainhash.Hash, int64, error) {
	blockHash, err := client.GetBlockHash(int64(height))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get block hash for height %d: %v", height, err)
	}
	return blockHash, int64(height), nil
}

// getTransactionFromBlock fetches a block (verbosity=2) and scans for the transaction by txid.
func getTransactionFromBlock(client *rpcclient.Client, txid string, blockHash *chainhash.Hash) (btcjson.TxRawResult, int64, error) {
	rawBlock, err := client.RawRequest("getblock", []json.RawMessage{
		json.RawMessage(fmt.Sprintf(`"%s"`, blockHash.String())),
		json.RawMessage("2"),
	})
	if err != nil {
		return btcjson.TxRawResult{}, 0, fmt.Errorf("failed to fetch block %s: %v", blockHash, err)
	}

	var blockResult btcjson.GetBlockVerboseTxResult
	if err := json.Unmarshal(rawBlock, &blockResult); err != nil {
		return btcjson.TxRawResult{}, 0, fmt.Errorf("failed to parse block %s: %v", blockHash, err)
	}

	for _, tx := range blockResult.Tx {
		if tx.Txid == txid {
			tx.BlockHash = blockHash.String()
			return tx, blockResult.Height, nil
		}
	}

	return btcjson.TxRawResult{}, 0, fmt.Errorf("transaction %s not found in block %s", txid, blockHash)
}

// extractTransactionDetails retrieves the block height from the transaction's block header.
func extractTransactionDetails(client *rpcclient.Client, txVerbose *btcjson.TxRawResult) (Transaction, int64, error) {
	var blockHeight int64
	if txVerbose.BlockHash != "" {
		blockHash, err := chainhash.NewHashFromStr(txVerbose.BlockHash)
		if err != nil {
			return Transaction{}, 0, fmt.Errorf("invalid block hash: %v", err)
		}
		header, err := client.GetBlockHeaderVerbose(blockHash)
		if err != nil {
			return Transaction{}, 0, fmt.Errorf("getblockheader failed: %v", err)
		}
		blockHeight = int64(header.Height)
	}
	tx := Transaction{
		TxID:      txVerbose.Txid,
		BlockHash: txVerbose.BlockHash,
	}
	return tx, blockHeight, nil
}

// extractAddress extracts a Bitcoin address from a PkScript.
func extractAddress(pkScript []byte) (string, error) {
	_, addresses, _, err := txscript.ExtractPkScriptAddrs(pkScript, &chaincfg.TestNet3Params)
	if err != nil || len(addresses) == 0 {
		return "", fmt.Errorf("failed to extract address: %v", err)
	}
	return addresses[0].String(), nil
}

// marshalJSON is a helper to marshal a value to JSON.
func marshalJSON(v interface{}) string {
	bytes, err := json.Marshal(v)
	if err != nil {
		fmt.Printf("Error marshalling JSON: %v\n", err)
		return ""
	}
	return string(bytes)
}

// continuousUTXOMonitor polls for new blocks and processes them.
func continuousUTXOMonitor(ctx context.Context, client *rpcclient.Client, cache *Cache) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			currentBlock, err := client.GetBlockCount()
			if err != nil {
				fmt.Printf("Failed to get block count: %v\n", err)
				continue
			}
			lastBlock := cache.getBlockHeight()
			if currentBlock > lastBlock {
				fmt.Printf("New block detected: %d -> %d, updating UTXOs...\n", lastBlock, currentBlock)
				processNewBlock(client, currentBlock, cache)
				cache.updateBlockHeight(currentBlock)
			}
		}
	}
}

// processNewBlock processes a new block and caches UTXOs for watched addresses.
func processNewBlock(client *rpcclient.Client, blockHeight int64, cache *Cache) {
	blockHash, err := client.GetBlockHash(blockHeight)
	if err != nil {
		fmt.Printf("Failed to get block hash: %v\n", err)
		return
	}

	block, err := client.GetBlock(blockHash)
	if err != nil {
		fmt.Printf("Failed to get block %s: %v\n", blockHash, err)
		return
	}

	for _, tx := range block.Transactions {
		txID := tx.TxHash().String()
		cache.addTransaction(Transaction{
			TxID:      txID,
			BlockHash: blockHash.String(),
		})

		for i, output := range tx.TxOut {
			address, err := extractAddress(output.PkScript)
			if err != nil || address == "" {
				continue
			}
			if isWatchedAddress(address) {
				fmt.Printf("Found UTXO for address %s in tx %s, Vout: %d, Amount: %.8f\n",
					address, txID, i, float64(output.Value)/100000000)
				utxo := UTXO{
					TxID:         txID,
					Vout:         i,
					Amount:       float64(output.Value) / 100000000,
					ScriptPubKey: fmt.Sprintf("addr(%s)", address),
				}
				cache.addUTXO(address, utxo)
			}
		}
	}
	fmt.Printf("Indexed UTXOs from block %d\n", blockHeight)
}

// isWatchedAddress checks if an address is in the watched list.
func isWatchedAddress(address string) bool {
	for _, addr := range watchedAddresses {
		if addr == address {
			return true
		}
	}
	return false
}
