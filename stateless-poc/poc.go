package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
)

// Testnet RPC Config
const (
	rpcHost = "localhost:48332" // Testnet default RPC port
	rpcUser = "admin"
	rpcPass = "admin"
)

// Wallet and UTXO Details (Manually Extracted from listunspent)
const (

	// Recipient address
	// This is the address that will receive the funds
	recipient = "mt7Wd4k9KSs6f7XtAZY96JTsPfxmZLWNMN"

	/*
		Please read the README.md file to understand how to extract the UTXO details
	*/

	// Address that owns the UTXO
	addressStr = "muCmmr3fwCvbFbdPUgtw6KFyx92qtDyuyx"

	privateKeyWIF = "XXXXXXXXXXXXXXXXXXXXXXXXXXXXX"

	// UTXO details (we use the one with enough funds)
	utxoTxID     = "ab9a228797321d42486449895a5689a353eee963006344526e2f0c85fce2bf0a"
	utxoVout     = 0
	scriptPubKey = "76a91496217dc748df395162630a1692fa685b4d66e44188ac"
	utxoAmount   = 14872 // 0.00014872 BTC in Satoshis

	// Transaction details
	amountToSend = 10000 // Sending 10,000 Satoshis (0.0001 BTC)
	fee          = 1000  // Transaction fee (1,000 Satoshis)
)

// Connect to Bitcoin RPC
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

// Create and sign a transaction manually
func createAndSignTx() (string, error) {
	wif, err := btcutil.DecodeWIF(privateKeyWIF)
	if err != nil {
		return "", fmt.Errorf("error decoding WIF: %w", err)
	}

	// Create a new transaction
	tx := wire.NewMsgTx(wire.TxVersion)

	// Parse tx hash
	txHash, err := chainhash.NewHashFromStr(utxoTxID)
	if err != nil {
		return "", err
	}

	// Create outpoint
	outPoint := wire.NewOutPoint(txHash, utxoVout)

	// Create and add input (spending the UTXO)
	txIn := wire.NewTxIn(outPoint, nil, nil)
	tx.AddTxIn(txIn)

	// Create recipient output
	recipientAddr, err := btcutil.DecodeAddress(recipient, &chaincfg.TestNet4Params)
	if err != nil {
		return "", err
	}
	pkScript, err := txscript.PayToAddrScript(recipientAddr)
	if err != nil {
		return "", err
	}
	tx.AddTxOut(wire.NewTxOut(amountToSend, pkScript))

	// Add change output (if needed)
	change := utxoAmount - (amountToSend + fee)

	if change > 546 { // If change is too small, it will be eaten by fees
		changeAddr, err := btcutil.DecodeAddress(addressStr, &chaincfg.TestNet4Params)
		if err != nil {
			return "", err
		}
		changeScript, err := txscript.PayToAddrScript(changeAddr)
		if err != nil {
			return "", err
		}
		tx.AddTxOut(wire.NewTxOut(int64(change), changeScript))

		fmt.Println("Added Change Output:")
		fmt.Printf("  - Address: %s\n", addressStr)
		fmt.Printf("  - Change Amount: %d Satoshis\n", change)

	} else {
		fmt.Println("Change too small, adding it to fee")
	}

	// Decode scriptPubKey
	scriptPubKeyBytes, _ := hex.DecodeString(scriptPubKey)

	// Sign the transaction
	for i, txIn := range tx.TxIn {
		sigScript, err := txscript.SignatureScript(tx, i, scriptPubKeyBytes, txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			return "", err
		}
		txIn.SignatureScript = sigScript
	}

	// Serialize and return hex
	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return "", err
	}
	signedTxHex := hex.EncodeToString(buf.Bytes())

	return signedTxHex, nil
}

// Broadcast transaction to Bitcoin network
func broadcastTx(client *rpcclient.Client, signedTxHex string) error {
	rawTxBytes, err := hex.DecodeString(signedTxHex)
	if err != nil {
		return err
	}

	tx := wire.NewMsgTx(wire.TxVersion)
	err = tx.Deserialize(bytes.NewReader(rawTxBytes))
	if err != nil {
		return err
	}

	txHash, err := client.SendRawTransaction(tx, false)
	if err != nil {
		return err
	}
	fmt.Printf("Transaction submitted: %s\n", txHash.String())
	return nil
}

func main() {
	// Connect to Bitcoin node
	client, err := connectRPC()
	if err != nil {
		log.Fatalf("Error connecting to Bitcoin RPC: %v", err)
	}
	defer client.Shutdown()

	// Create and sign transaction
	signedTxHex, err := createAndSignTx()
	if err != nil {
		log.Fatalf("Error creating and signing transaction: %v", err)
	}

	// Print the signed transaction (for debugging)
	fmt.Println("Signed Transaction Hex:", signedTxHex)

	// Broadcast transaction
	err = broadcastTx(client, signedTxHex)
	if err != nil {
		log.Fatalf("Error broadcasting transaction: %v", err)
	}
}
