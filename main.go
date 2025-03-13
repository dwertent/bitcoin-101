package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type hostURL string

const (
	walletURL hostURL = "https://127.0.0.1:48331/"
	nodeURL   hostURL = "http://127.0.0.1:48332/"
)

// RPCRequest defines the JSON-RPC request structure.
type RPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// RPCResponse defines the JSON-RPC response structure.
type RPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Result  interface{} `json:"result"`
	Error   interface{} `json:"error"`
}

// getBalance queries the wallet balance for the given account (or "*" for all addresses).
func getBalance(accountName string, confirmations int) (float64, error) {
	if accountName == "" {
		accountName = "*"
	}
	reqBody := RPCRequest{
		JSONRPC: "1.0",
		ID:      "goClientTest",
		Method:  "getbalance",
		Params:  []interface{}{accountName, confirmations},
	}
	rpcResp, err := sendRPCRequest(reqBody, walletURL)
	if err != nil {
		return 0, err
	}
	if rpcResp.Error != nil {
		return 0, fmt.Errorf("rpc error: %v", rpcResp.Error)
	}
	return rpcResp.Result.(float64), nil
}

// sendToAddress sends the specified amount to the given destination address.
func sendToAddress(address string, amount float64) (string, error) {
	reqBody := RPCRequest{
		JSONRPC: "1.0",
		ID:      "goClientTest",
		Method:  "sendtoaddress",
		Params:  []interface{}{address, amount},
	}
	rpcResp, err := sendRPCRequest(reqBody, walletURL)
	if err != nil {
		return "", err
	}
	if rpcResp.Error != nil {
		return "", fmt.Errorf("rpc error: %v", rpcResp.Error)
	}
	txid, ok := rpcResp.Result.(string)
	if !ok {
		return "", fmt.Errorf("unexpected result type: %T", rpcResp.Result)
	}
	return txid, nil
}

// gettransaction retrieves detailed information about a transaction.
// The "verbose" flag is set to true so that we receive a JSON object with details.
func getTransaction(txid string, verbose bool) (map[string]interface{}, error) {
	reqBody := RPCRequest{
		JSONRPC: "1.0",
		ID:      "gettransaction",
		Method:  "gettransaction",
		Params:  []interface{}{txid, verbose},
	}
	rpcResp, err := sendRPCRequest(reqBody, walletURL)
	if err != nil {
		return nil, err
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error: %v", rpcResp.Error)
	}
	result, ok := rpcResp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type for getrawtransaction result")
	}
	return result, nil
}

// getRawTransaction retrieves detailed information about a transaction.
// The "verbose" flag is set to true so that we receive a JSON object with details.
func getRawTransaction(txid, blockHash string, verbose bool) (map[string]interface{}, error) {
	reqBody := RPCRequest{
		JSONRPC: "1.0",
		ID:      "getRawTransaction",
		Method:  "getrawtransaction",
		Params:  []interface{}{txid, verbose},
	}
	if blockHash != "" {
		reqBody.Params = append(reqBody.Params, blockHash)
	}

	rpcResp, err := sendRPCRequest(reqBody, nodeURL)
	if err != nil {
		return nil, err
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error: %v", rpcResp.Error)
	}
	result, ok := rpcResp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type for getrawtransaction result")
	}
	return result, nil
}

// getBlock retrieves detailed information about a block given its hash.
func getBlock(blockHash string) (map[string]interface{}, error) {
	reqBody := RPCRequest{
		JSONRPC: "1.0",
		ID:      "getBlock",
		Method:  "getblock",
		Params:  []interface{}{blockHash},
	}
	rpcResp, err := sendRPCRequest(reqBody, nodeURL)
	if err != nil {
		return nil, err
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error: %v", rpcResp.Error)
	}
	result, ok := rpcResp.Result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected type for getblock result")
	}
	return result, nil
}

func unlockWallet() error {
	reqBody := RPCRequest{
		JSONRPC: "1.0",
		ID:      "goClientTest",
		Method:  "walletpassphrase",
		Params:  []interface{}{"admin", 10},
	}
	rpcResp, err := sendRPCRequest(reqBody, walletURL)
	if err != nil {
		return err
	}
	if rpcResp.Error != nil {
		return fmt.Errorf("rpc error: %v", rpcResp.Error)
	}
	return nil
}

// getNewAddress generates a new address for your wallet.
func getNewAddress(tag string) (string, error) {
	reqBody := RPCRequest{
		JSONRPC: "1.0",
		ID:      "goClientTest",
		Method:  "getnewaddress",
		Params:  []interface{}{},
	}
	if tag != "" {
		reqBody.Params = []interface{}{tag}
	}
	rpcResp, err := sendRPCRequest(reqBody, walletURL)
	if err != nil {
		return "", err
	}
	if rpcResp.Error != nil {
		return "", fmt.Errorf("rpc error: %v", rpcResp.Error)
	}
	address, ok := rpcResp.Result.(string)
	if !ok {
		return "", fmt.Errorf("unexpected result type: %T", rpcResp.Result)
	}
	return address, nil
}

// getAddressesByAccount retrieves all addresses for a specific account.
func getAddressesByAccount(accountName string) ([]string, error) {
	reqBody := RPCRequest{
		JSONRPC: "1.0",
		ID:      "goClientTest",
		Method:  "getaddressesbyaccount",
		Params:  []interface{}{},
	}
	if accountName != "" {
		reqBody.Params = []interface{}{accountName}
	}
	rpcResp, err := sendRPCRequest(reqBody, walletURL)
	if err != nil {
		return nil, err
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error: %v", rpcResp.Error)
	}
	addresses, ok := rpcResp.Result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rpcResp.Result)
	}
	var result []string
	for _, addr := range addresses {
		result = append(result, addr.(string))
	}
	return result, nil
}

func createWallet(walletName string) error {
	reqBody := RPCRequest{
		JSONRPC: "1.0",
		ID:      "goClientTest",
		Method:  "createwallet",
		Params:  []interface{}{walletName},
	}
	rpcResp, err := sendRPCRequest(reqBody, walletURL)
	if err != nil {
		return err
	}
	if rpcResp.Error != nil {
		return fmt.Errorf("rpc error: %v", rpcResp.Error)
	}
	return nil
}

func listUnspent(address string) ([]interface{}, error) {
	reqBody := RPCRequest{
		JSONRPC: "1.0",
		ID:      "goClientTest",
		Method:  "listunspent",
		Params:  []interface{}{},
	}
	if address != "" {
		reqBody.Params = []interface{}{0, 9999999, []string{address}}
	}
	rpcResp, err := sendRPCRequest(reqBody, walletURL)
	if err != nil {
		return nil, err
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error: %v", rpcResp.Error)
	}
	unspent, ok := rpcResp.Result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", rpcResp.Result)
	}
	return unspent, nil
}

func dumpPrivKey(address string) (string, error) {
	if err := walletPassphrase(); err != nil {
		return "", err
	}

	reqBody := RPCRequest{
		JSONRPC: "1.0",
		ID:      "goClientTest",
		Method:  "dumpprivkey",
		Params:  []interface{}{address},
	}
	rpcResp, err := sendRPCRequest(reqBody, walletURL)
	if err != nil {
		return "", err
	}
	if rpcResp.Error != nil {
		return "", fmt.Errorf("rpc error: %v", rpcResp.Error)
	}
	key, ok := rpcResp.Result.(string)
	if !ok {
		return "", fmt.Errorf("unexpected result type: %T", rpcResp.Result)
	}
	return key, nil
}

func walletPassphrase() error {
	passphrase := "admin"
	timeout := 10
	reqBody := RPCRequest{
		JSONRPC: "1.0",
		ID:      "goClientTest",
		Method:  "walletpassphrase",
		Params:  []interface{}{passphrase, timeout},
	}
	rpcResp, err := sendRPCRequest(reqBody, walletURL)
	if err != nil {
		return err
	}
	if rpcResp.Error != nil {
		return fmt.Errorf("rpc error: %v", rpcResp.Error)
	}
	return nil
}

// sendRPCRequest sends a JSON-RPC request to your wallet node.
func sendRPCRequest(reqBody RPCRequest, host hostURL) (RPCResponse, error) {
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return RPCResponse{}, err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // WARNING: Only for testing.
		},
	}

	client := &http.Client{Transport: tr}
	request, err := http.NewRequest("POST", string(host), bytes.NewBuffer(reqBytes))
	if err != nil {
		return RPCResponse{}, err
	}

	request.SetBasicAuth("admin", "admin")
	request.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(request)
	if err != nil {
		return RPCResponse{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return RPCResponse{}, err
	}
	var rpcResp RPCResponse
	err = json.Unmarshal(body, &rpcResp)
	if err != nil {
		return RPCResponse{}, err
	}
	return rpcResp, nil
}

func main() {
	switch os.Args[1] {

	case "createwallet":
		// Create a new wallet.
		if len(os.Args) < 3 {
			panic("Usage: go run main.go createwallet <walletname>")
		}
		walletName := os.Args[2]
		err := createWallet(walletName)
		if err != nil {
			panic(err)
		}

	case "getaddressesbyaccount":
		// Get all addresses for a specific account.
		if len(os.Args) < 3 {
			panic("Usage: go run main.go getaddressesbyaccount <account>")
		}
		accountName := os.Args[2]
		address, err := getAddressesByAccount(accountName)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Addresses for account '%s':\n", accountName)
		fmt.Printf("%s\n", strings.Join(address, ", "))

	case "getnewaddress":
		// Get a new address for your wallet.
		tag := ""
		if len(os.Args) >= 3 {
			tag = os.Args[2]
		}
		address, err := getNewAddress(tag)
		if err != nil {
			panic(err)
		}
		fmt.Printf("New address: %s\n", address)

	case "getbalance":
		accountName := ""
		if len(os.Args) >= 3 {
			accountName = os.Args[2]
		}

		// default to 1 confirmation - the transaction was included submitted to the chain (resent block).
		// if set to 0 - block was not added to the chain yet. (funds are not available for spending)
		confirmations := 1
		if len(os.Args) >= 4 {
			if t, err := strconv.Atoi(os.Args[3]); err == nil {
				confirmations = t
			}
		}

		// Get your current balance (from all addresses by default).
		balance, err := getBalance(accountName, confirmations)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Balance: %f BTC\n", balance)

	case "send":
		if len(os.Args) < 3 {
			panic("Usage: go run main.go send <destination> [amount]")
		}
		dest := os.Args[2]

		amount := 0.00000546 // minimum amount allowed to send
		if len(os.Args) >= 4 {
			if tmp, err := strconv.ParseFloat(os.Args[3], 64); err == nil {
				amount = tmp
			}
		}

		// Calculate the amount to send: 1/36 of your balance.
		fmt.Printf("Sending %f BTC to %s\n", amount, dest)

		if err := unlockWallet(); err != nil {
			panic(err)
		}

		// Send the funds.
		txid, err := sendToAddress(dest, amount)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Transaction sent! TXID: %s\n", txid)

	// getrawtx: get raw transaction from the chain - this will include transactions that where not included in the block yet.
	case "getrawtx":

		if len(os.Args) < 3 {
			panic("Usage: go run main.go getrawtx <txid> [blockhash]")
		}
		txid := os.Args[2]

		blockHash := ""
		if len(os.Args) >= 4 {
			blockHash = os.Args[3]
		}

		fmt.Printf("Getting details for transaction %s\n", txid)

		// Example: Get details about a specific transaction.
		// Replace this txid with one you wish to inspect. Here we use a sample txid.
		txDetails, err := getRawTransaction(txid, blockHash, true)
		if err != nil {
			if strings.Contains(err.Error(), "No such mempool transaction") {
				fmt.Println("Transaction not found in the mempool, run getrawtx with the blockhash to get the transaction details.")
				os.Exit(1)
			}
			panic(err)
		}
		b, err := json.Marshal(txDetails)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Transaction Details:\n%s\n", b)

	// gettx: get transaction from the wallet - this will include transactions that where included in the block. and that are associated with the wallet.
	case "gettx":
		if len(os.Args) < 3 {
			panic("Usage: go run main.go gettx <txid>")
		}
		txid := os.Args[2]

		fmt.Printf("Getting details for transaction %s\n", txid)

		// Example: Get details about a specific transaction.
		// Replace this txid with one you wish to inspect. Here we use a sample txid.
		txDetails, err := getTransaction(txid, true)
		if err != nil {
			panic(err)
		}
		b, err := json.Marshal(txDetails)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Transaction Details:\n%s\n", b)

	case "getblock":
		if len(os.Args) < 3 {
			panic("Usage: go run main.go getblock <blockhash>")
		}
		blockHash := os.Args[2]
		block, err := getBlock(blockHash)
		if err != nil {
			panic(err)
		}
		block["tx"] = []string{"<list of txIDs>"}
		b, err := json.Marshal(block)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Block Details for block %s:\n%s\n", blockHash, b)

	case "listunspent":
		// address input (optional)
		address := ""
		if len(os.Args) >= 3 {
			address = os.Args[2]
		}
		unspent, err := listUnspent(address)
		if err != nil {
			panic(err)
		}
		b, err := json.Marshal(unspent)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Unspent Transactions:\n%s\n", b)

	case "dumpprivkey":
		if len(os.Args) < 3 {
			panic("Usage: go run main.go dumpprivkey <address>")
		}
		address := os.Args[2]

		key, err := dumpPrivKey(address)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Private key for address: %s\n", key)

	default:
		fmt.Println("Invalid command. Please use one of the following:")
		fmt.Println("getaddressesbyaccount <account>")
		fmt.Println("getnewaddress [tag]")
		fmt.Println("getbalance [account] [confirmations]")
		fmt.Println("send <destination> [amount]")
		fmt.Println("getrawtx <txid>")
		fmt.Println("gettx <txid>")
		fmt.Println("getblock <blockhash>")
		fmt.Println("listunspent [address]")
		fmt.Println("dumpprivkey <address>")
	}
}
