# Overview

**For an in-depth understanding of Bitcoin fundamentals, please refer to the [Bitcoin 101 Guide](Bitcoin-101.md).**

The purpose of this guide is not to replicate the API but rather to make it easy to explore the API by running some basic commands against the wallet and the blockchain.

There are more APIs available, but many are not implemented here. This guide covers the basics, such as creating new addresses and transferring money between them.

You can read more about the built-in RPC APIs here: [Bitcoin RPC API Reference](https://developer.bitcoin.org/reference/rpc/).

This example connects to the `btcwallet` project. Unfortunately, not all APIs are implemented.

---

## Prerequisites

- A running blockchain node
- A running wallet

### Setting Up the Environment

#### Install `bitcoind`
```sh
brew install bitcoind
```

#### Run `bitcoind`
```sh
/opt/homebrew/bin/bitcoind -testnet4 -disablewallet -rpcuser=admin -rpcpassword=admin
```

> **Note:** This will take a while as the node syncs with the blockchain.

Alternatively, you can run the Bitcoin node with the wallet enabled. In such a case, there is no need to install `btcwallet` separately. To do this, run:
```sh
/opt/homebrew/bin/bitcoind -testnet4 -rpcuser=admin -rpcpassword=admin
```

If you choose this method, you can use the built-in wallet functionality of `bitcoind` instead of setting up `btcwallet`.

#### Clone `btcwallet`
```sh
git clone https://github.com/dwertent/btcwallet.git -b feat/support-testnet4
cd btcwallet
```

#### Create a Wallet
```sh
GO111MODULE=on go run . ./cmd/... -u admin -P admin --testnet4 --backend=bitcoind --create
```

> **For convenience, use 'admin' as the wallet password.**

#### Run the Wallet
```sh
GO111MODULE=on go run . ./cmd/... -u admin -P admin --testnet4 --backend=bitcoind
```

You should see logs similar to:
```sh
[INF] WLLT: Opened wallet
[INF] BTCW: Connected successfully to a bitcoind node.
```

---

## Generate an Address

#### Clone the `bitcoin-101` Repository
```sh
git clone https://github.com/dwertent/bitcoin-101
cd bitcoin-101
```

#### Run the Following Command
```sh
go run . getnewaddress
```

The output should be a new Bitcoin address.

Alternatively, you can generate the address by running an RPC request:
```sh
curl --user admin:admin --data-binary '{"jsonrpc": "1.0", "id": "curltest", "method": "getnewaddress", "params": []}' -k https://127.0.0.1:48331/
```

---

## Fund the Account
Go to [Bitcoin Testnet4 Faucet](https://coinfaucet.eu/en/btc-testnet4/) and paste the generated address to receive testnet coins.

> **Note:** It may take some time for the transaction to be confirmed on the blockchain.

---

## View the BoltDB

#### Install BoltDB Package
```sh
go get github.com/br0xen/boltbrowser
```

#### Run BoltDB Browser
```sh
boltbrowser ~/Library/Application\ Support/Btcwallet/testnet4/wallet.db
```

This allows you to inspect the wallet database.

---

This guide should help you set up and interact with a basic Bitcoin wallet on Testnet4.

