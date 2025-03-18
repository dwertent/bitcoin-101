package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
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
	addressStr    = "muCmmr3fwCvbFbdPUgtw6KFyx92qtDyuyx"
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

func prepareTx() (*wire.MsgTx, error) {
	tx := wire.NewMsgTx(wire.TxVersion)
	txHash, err := chainhash.NewHashFromStr(utxoTxID)
	if err != nil {
		return nil, err
	}
	outPoint := wire.NewOutPoint(txHash, utxoVout)
	tx.AddTxIn(wire.NewTxIn(outPoint, nil, nil))

	recipientAddr, err := btcutil.DecodeAddress(recipient, &chaincfg.TestNet4Params)
	if err != nil {
		return nil, err
	}
	pkScript, err := txscript.PayToAddrScript(recipientAddr)
	if err != nil {
		return nil, err
	}
	tx.AddTxOut(wire.NewTxOut(amountToSend, pkScript))

	change := utxoAmount - (amountToSend + fee)
	if change > 546 {
		changeAddr, err := btcutil.DecodeAddress(addressStr, &chaincfg.TestNet4Params)
		if err != nil {
			return nil, err
		}
		changeScript, err := txscript.PayToAddrScript(changeAddr)
		if err != nil {
			return nil, err
		}
		tx.AddTxOut(wire.NewTxOut(int64(change), changeScript))
	}

	m, _ := json.Marshal(tx)
	fmt.Printf("Prepared Transaction:\n%s\n", m)

	return tx, nil
}

func signTx(tx *wire.MsgTx) (string, error) {
	wif, err := btcutil.DecodeWIF(privateKeyWIF)
	if err != nil {
		return "", fmt.Errorf("error decoding WIF: %w", err)
	}
	scriptPubKeyBytes, _ := hex.DecodeString(scriptPubKey)
	for i, txIn := range tx.TxIn {
		sigScript, err := txscript.SignatureScript(tx, i, scriptPubKeyBytes, txscript.SigHashAll, wif.PrivKey, true)
		if err != nil {
			return "", err
		}
		txIn.SignatureScript = sigScript
	}

	var buf bytes.Buffer
	if err := tx.Serialize(&buf); err != nil {
		return "", err
	}
	signedTxHex := hex.EncodeToString(buf.Bytes())
	fmt.Println("Signed Transaction Hex:", signedTxHex)
	return signedTxHex, nil
}

func broadcastTx(client *rpcclient.Client, signedTxHex string) error {
	rawTxBytes, err := hex.DecodeString(signedTxHex)
	if err != nil {
		return err
	}

	tx := wire.NewMsgTx(wire.TxVersion)
	if err := tx.Deserialize(bytes.NewReader(rawTxBytes)); err != nil {
		return err
	}

	jsonTx, _ := json.Marshal(tx)
	fmt.Printf("Broadcasting Transaction:\n%s\n", jsonTx)

	txHash, err := client.SendRawTransaction(tx, false)
	if err != nil {
		return err
	}
	fmt.Printf("Transaction submitted: %s\n", txHash.String())
	return nil
}

func main() {
	client, err := connectRPC()
	if err != nil {
		log.Fatalf("Error connecting to Bitcoin RPC: %v", err)
	}
	defer client.Shutdown()

	tx, err := prepareTx()
	if err != nil {
		log.Fatalf("Error preparing transaction: %v", err)
	}

	signedTxHex, err := signTx(tx)
	if err != nil {
		log.Fatalf("Error signing transaction: %v", err)
	}

	fmt.Println("Final Signed Transaction:", signedTxHex)

	if err := broadcastTx(client, signedTxHex); err != nil {
		log.Fatalf("Error broadcasting transaction: %v", err)
	}
}
