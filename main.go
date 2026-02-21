// Package main is the entry point for the blockchain demo.
// It creates a chain, adds blocks, then prints each block and verifies its proof-of-work.
package main

import (
	"fmt"
	"go-blockchain/block"
	"go-blockchain/work"
	"strconv"
)

func main() {
	// Create a new blockchain. This automatically creates the genesis (first) block
	// so the chain always has at least one block to link from.
	bc := block.NewBlockchain()

	// Add two blocks with sample transaction data. Each block is mined (PoW) before
	// being appended; mining may take a while depending on targetBits.
	bc.AddBlock("Send 1 BTC to Ivan")
	bc.AddBlock("Send 2 more BTC to Ivan")

	// Iterate over all blocks and print their contents, then verify proof-of-work.
	// We build work.BlockData from each block because the work package cannot
	// import block (that would create an import cycle), so we pass only the
	// fields needed for hashing and validation.
	for _, block := range bc.Blocks() {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := work.NewProofOfWork(&work.BlockData{
			PrevBlockHash: block.PrevBlockHash,
			Data:          block.Data,
			Timestamp:     block.Timestamp,
			Nonce:         block.Nonce,
		})
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
