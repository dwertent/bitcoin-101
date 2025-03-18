
# Bitcoin UTXO Indexer Proof-of-Concept (POC)

This POC demonstrates how to index UTXOs for a set of watched Bitcoin addresses by interacting with a Bitcoin node via RPC. The indexer performs an initial full scan using the `scantxoutset` RPC call and then continuously monitors the blockchain for new blocks to update its cache.

## Overview

The POC code connects to a Bitcoin Testnet node and uses RPC calls to:

- **Scan for UTXOs:**  
  It invokes the `scantxoutset` RPC command for a predefined list of addresses. The response includes details about unspent outputs such as the transaction ID, output index (vout), amount, locking script (scriptPubKey), and confirmation block height.

- **Retrieve Transaction Details:**  
  For each UTXO found, the indexer tries to obtain additional transaction data. It first attempts a direct lookup (which works if the transaction is still in the mempool or part of the wallet). If that fails, it uses the confirmation height to fetch the corresponding block hash and then retrieves the transaction from the full block (fetched with verbosity level 2).

- **Cache Results:**  
  A simple in-memory cache stores UTXOs and basic transaction information. The cache is keyed by watched addresses and includes minimal transaction details (TxID and BlockHash) along with the block height.

- **Continuous Monitoring:**  
  A background goroutine polls the node every 10 seconds to check for new blocks. If a new block is detected (i.e. the block count increases), the indexer processes that block to update its cache with any new UTXOs for the watched addresses.

## Code Components

### Data Structures

- **UTXO Structure:**  
  Represents an unspent output with the following fields:
  - `TxID`: The transaction identifier.
  - `Vout`: Output index within the transaction.
  - `Amount`: Amount (in BTC) held by the UTXO.
  - `ScriptPubKey`: Locking script in hexadecimal format.
  - `Height`: Block height where the UTXO was confirmed.
  - `Desc`: An optional description field.

- **Transaction Structure:**  
  Holds basic transaction information:
  - `TxID`: Transaction identifier.
  - `BlockHash`: Hash of the block that includes the transaction.
  - `Inputs` and `Outputs`: (Currently, only outputs are stored in the UTXO cache.)
 

## Flow 

1. **Full Scan:**  
   It runs `startFullScan` to perform an initial full UTXO scan using `scantxoutset` for the watched addresses. For each UTXO, the code attempts to retrieve additional transaction details and caches both the UTXO and minimal transaction data.

2. **Continuous Monitoring:**  
   A background goroutine (`continuousUTXOMonitor`) checks for new blocks every 10 seconds. When a new block is detected, it processes the block to identify any new UTXOs relevant to the watched addresses and updates the cache accordingly.

## Usage

1. **Start the Bitcoin Testnet node** 
2. **Compile and run the indexer:**
   ```sh
   go run .
   ```
3. **Observe the logs:**  
   The console output will display messages indicating:
   - The start of the UTXO scan.
   - UTXOs found for each watched address.
   - Any transaction details retrieved.
   - New blocks detected and processed.
4. **Shutdown Reporting:**  
   When the program exits, it prints a summary list of all UTXOs and transactions cached during its operation.
 