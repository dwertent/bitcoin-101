# Bitcoin Transaction POC

## Overview
This is a simple guide to help you manually create, sign, and send Bitcoin transactions on **Testnet4**. It's designed for cases where you already have a service that tracks UTXOs and another service that handles private keys.

### What This Script Does **Not** Do
- **It’s NOT an indexer** – It doesn’t scan the blockchain for UTXOs. You need to provide them.
- **It doesn’t store private keys** – Keys come from a separate key management service.

## Setting Up Your Transaction

### **Start Your Testnet4 Node**
Before anything, make sure you have a **Bitcoin Testnet4** node running with a wallet.
Follow the [installation and wallet setup guide](../README.md)

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

## **APIs**
### **prepareTx**
**Input:**
- `utxoTxID` (string) - The transaction ID of the UTXO.
- `utxoVout` (int) - The output index in the transaction.
- `scriptPubKey` (string) - The script required to unlock the UTXO.
- `utxoAmount` (int) - The amount of Bitcoin in Satoshis.
- `recipient` (string) - The address receiving the funds.
- `amountToSend` (int) - The amount of Satoshis to send.
- `fee` (int) - The transaction fee in Satoshis.

**Output:**
- `*wire.MsgTx` - The prepared transaction.
- `error` - Any errors encountered during preparation.

### **signTx**
**Input:**
- `tx` (*wire.MsgTx) - The prepared transaction.
- `privateKeyWIF` (string) - The private key in Wallet Import Format (WIF).
- `scriptPubKey` (string) - The scriptPubKey required to unlock the UTXO.

**Output:**
- `string` - The signed transaction in hex format.
- `error` - Any errors encountered during signing.

### **broadcastTx**
**Input:**
- `client` (*rpcclient.Client) - The Bitcoin RPC client.
- `signedTxHex` (string) - The signed transaction in hex format.

**Output:**
- `error` - Any errors encountered while broadcasting.

### **Example Execution Logs**

#### Prepared Transaction:
```
{
  "Version": 1,
  "TxIn": [
    {
      "PreviousOutPoint": {
        "Hash": "ab9a228797321d42486449895a5689a353eee963006344526e2f0c85fce2bf0a",
        "Index": 0
      },
      "SignatureScript": null,
      "Witness": null,
      "Sequence": 4294967295
    }
  ],
  "TxOut": [
    {
      "Value": 10000,
      "PkScript": "dqkUiiqjUp6IWxYaV8sg8tB9zKfTDp+IrA=="
    },
    {
      "Value": 3872,
      "PkScript": "dqkUliF9x0jfOVFiYwoWkvpoW01m5EGIrA=="
    }
  ],
  "LockTime": 0
}
```

#### Signed Transaction:
```
01000000010abfe2fc850c2f6e5244630063e9ee53a389565a89496448421d329787229aab000000006a473044...
```


#### Transaction to submit:
```
{
  "Version": 1,
  "TxIn": [
    {
      "PreviousOutPoint": {
        "Hash": "ab9a228797321d42486449895a5689a353eee963006344526e2f0c85fce2bf0a",
        "Index": 0
      },
      "SignatureScript": "RzBEAiA0lWs4sKqdJSoz8vAhZFp+dCP4A+rTkmekYFZOSi5DJgIgRc806iCJfgf9BaVyiQVwP/LzIbam7lwRIgzaBvK3V/MBIQKcSEk2t2TrfBzbgxXv92+qOKosC2VdYUsc/BnzajwMtA==",
      "Witness": null,
      "Sequence": 4294967295
    }
  ],
  "TxOut": [
    {
      "Value": 10000,
      "PkScript": "dqkUiiqjUp6IWxYaV8sg8tB9zKfTDp+IrA=="
    },
    {
      "Value": 3872,
      "PkScript": "dqkUliF9x0jfOVFiYwoWkvpoW01m5EGIrA=="
    }
  ],
  "LockTime": 0
}
```

#### Transaction submitted:
```
2e07e9b16c0dacbd98e03370dc2d364444f1c309615593135915bb0d372ddd42
```

And here it is in the mempool:
👉 [View in Mempool](https://mempool.space/testnet4/tx/2e07e9b16c0dacbd98e03370dc2d364444f1c309615593135915bb0d372ddd42)



 