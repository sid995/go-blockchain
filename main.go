// Package main is the entry point for the blockchain application.
// It wires the CLI to the blockchain and runs the command specified by the user.
// For createblockchain we do not open an existing DB (there is none yet); for
// all other commands we open blockchain.db and then run the CLI.
package main

import (
	"os"

	"go-blockchain/block"
	"go-blockchain/cli"
)

func main() {
	// No subcommand: show usage and exit. Run with nil bc so we never touch the DB.
	if len(os.Args) < 2 {
		c := cli.NewCLI(nil)
		c.Run()
		return
	}
	// createblockchain creates the DB and genesis block; no existing chain needed.
	// Pass nil bc; the createblockchain handler does not use it.
	if os.Args[1] == "createblockchain" {
		c := cli.NewCLI(nil)
		c.Run()
		return
	}
	// All other commands require an existing chain.
	bc := block.NewBlockchain("")
	defer bc.Close()
	c := cli.NewCLI(bc)
	c.Run()
}
