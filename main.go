package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

func getBalance(walletAdd string) (float64, error) {
	if walletAdd == "" {
		walletAdd = "*"
	}
	// Create the request payload.
	reqBody := RPCRequest{
		JSONRPC: "1.0",
		ID:      "goClientTest",
		Method:  "getbalance",
		Params:  []interface{}{walletAdd, 1},
	}
	rpcResp, err := sendRPCRequest(reqBody)
	if err != nil {
		return 0, err
	}
	return rpcResp.Result.(float64), nil
}

func sendRPCRequest(reqBody RPCRequest) (RPCResponse, error) {
	// Marshal the request into JSON.
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return RPCResponse{}, err
	}

	// Create a custom transport with TLS configuration that skips certificate verification.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // WARNING: Disables certificate verification.
		},
	}

	// Create a new HTTP client using the custom transport.
	client := &http.Client{Transport: tr}
	request, err := http.NewRequest("POST", "https://127.0.0.1:48331/", bytes.NewBuffer(reqBytes))
	if err != nil {
		return RPCResponse{}, err
	}

	// Set basic auth credentials.
	request.SetBasicAuth("admin", "admin")
	request.Header.Set("Content-Type", "application/json")

	// Send the request.
	resp, err := client.Do(request)
	if err != nil {
		return RPCResponse{}, err
	}
	defer resp.Body.Close()

	// Read and unmarshal the response.
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

	walletAdd := ""

	balance, err := getBalance(walletAdd)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Balance: %f\n", balance)

}
