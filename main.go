// Package main is the entry point for the blockchain demo.
// It creates or opens a persistent blockchain (BoltDB), then prints its representation.
package main

import (
	"fmt"
	"go-blockchain/block"
)

func main() {
	// NewBlockchain opens or creates the DB and ensures a genesis block exists.
	// The chain is stored in blockchain.db; repeated runs reuse the same chain.
	bc := block.NewBlockchain()
	fmt.Print(bc)
}
