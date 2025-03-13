# Bitcoin Transaction POC

## Overview
This is a simple guide to help you manually create, sign, and send Bitcoin transactions on **Testnet4**. It's designed for cases where you already have a service that tracks UTXOs and another service that handles private keys.

### What This Script Does **Not** Do
- **It’s NOT an indexer** – It doesn’t scan the blockchain for UTXOs. You need to provide them.
- **It doesn’t store private keys** – Keys come from a separate key management service.

## Setting Up Your Transaction

### **Start Your Testnet4 Node**
Before anything, make sure you have a **Bitcoin Testnet4** node running with a wallet.
Follow the [installation and wallet setup guid](../README.md)

### **Get Your UTXOs**
Run this command to see your available UTXOs:
```sh
$ ../go run . listunspent
```
This will list all unspent transactions in your wallet.

<details>

<summary>Example output</summary>

```json
[
  {
    "account": "default",
    "address": "mt7Wd4k9KSs6f7XtAZY96JTsPfxmZLWNMN",
    "amount": 0.00000546,
    "confirmations": 7,
    "scriptPubKey": "76a9148a2aa3529e885b161a57cb20f2d07dcca7d30e9f88ac",
    "spendable": true,
    "txid": "fa2010b1e58cbf275d88998d438799f36bc06e33512be914a0a615418079c3f2",
    "vout": 0
  },
  {
    "account": "default",
    "address": "mydBSdJF1fDfe34VJJ5v65cAtrm8w6QBW9",
    "amount": 0.00000546,
    "confirmations": 687,
    "scriptPubKey": "76a914c69fbc30551d7fc7b29c1be197c5452ab9a1c48088ac",
    "spendable": true,
    "txid": "ab9a228797321d42486449895a5689a353eee963006344526e2f0c85fce2bf0a",
    "vout": 1
  },
  {
    "account": "default",
    "address": "muCmmr3fwCvbFbdPUgtw6KFyx92qtDyuyx",
    "amount": 0.00014872,
    "confirmations": 687,
    "scriptPubKey": "76a91496217dc748df395162630a1692fa685b4d66e44188ac",
    "spendable": true,
    "txid": "ab9a228797321d42486449895a5689a353eee963006344526e2f0c85fce2bf0a",
    "vout": 0
  }
]
```

</details>


### **Pick a UTXO**
Find a UTXO that has enough Bitcoin for your transaction **plus fees**. Take note of:
* The **address** that owns it  
* The **TXID** (transaction ID)  
* The **VOUT** (output index)  
* The **scriptPubKey** (locking script)  
* The **amount** (in Satoshis)  

### **Get the Private Key**
Once you have a UTXO, get its private key:
```sh
$ ../go run . dumpprivkey <address>
```

> **Keep this private key secret!** It’s needed to sign the transaction.


### **Get a New Address for the Recipient**

If you need a new recipient address, run the following command:
```
$ go run . getnewaddress
```

This will generate a fresh Bitcoin Testnet address that you can use for receiving funds.

### **Fill in Your Transaction Details**
Update the script with your values:
```go
const (
    privateKeyWIF = "<Your-Private-Key-WIF>"
    recipient     = "<Recipient-Address>"

    addressStr    = "<Your-Bitcoin-Testnet-Address>"
    utxoTxID      = "<Transaction-ID>"
    utxoVout      = <UTXO-Vout>
    scriptPubKey  = "<UTXO-scriptPubKey>"
    utxoAmount    = <UTXO-Amount-in-Satoshis>

    amountToSend  = <Amount-to-Send-in-Satoshis>
    fee           = <Transaction-Fee-in-Satoshis>
)
```

### **Run the Script to Send the Transaction**
Once everything is set up, run:
```sh
$ go run .
```
If everything works, you’ll see a **transaction ID** that you can track online!
 

### **Confirm the Transaction was Sent**

- Check if your transaction is broadcasted on a **Testnet block explorer**:
  👉 [mempool](https://mempool.space/testnet4)
