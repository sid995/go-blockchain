// Package main is the entry point for the blockchain demo.
// It creates or opens a persistent blockchain (BoltDB), then prints its representation.
package main

import (
	"go-blockchain/block"
	"go-blockchain/cli"
)

func main() {
	// NewBlockchain opens or creates the DB and ensures a genesis block exists.
	// The chain is stored in blockchain.db; repeated runs reuse the same chain.
	bc := block.NewBlockchain()
	defer bc.Close()

	c := cli.NewCLI(bc)
	c.Run()
}
