// Package cli implements the command-line interface for the blockchain.
// It parses subcommands (addblock, printchain, getbalance, send) and delegates
// to the blockchain and transaction packages.
package cli

import (
	"flag"
	"fmt"
	"go-blockchain/block"
	T "go-blockchain/transaction"
	"go-blockchain/work"
	"log"
	"os"
	"strconv"
)

// CLI holds a reference to the blockchain so commands can add blocks,
// iterate, mine, etc. It is created with NewCLI and must be given a valid
// opened blockchain.
type CLI struct {
	bc *block.Blockchain
}

// NewCLI returns a CLI that will operate on the given blockchain. The caller
// is responsible for opening and closing the blockchain.
func NewCLI(bc *block.Blockchain) *CLI {
	return &CLI{bc: bc}
}

// printUsage prints the supported subcommands and flags to stdout.
func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  getbalance -address ADDRESS - Get balance of ADDRESS")
	fmt.Println("  createblockchain -address ADDRESS - Create a blockchain and send genesis block reward to ADDRESS")
	fmt.Println("  printchain - Print all the blocks of the blockchain")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT - Send AMOUNT of coins from FROM address to TO")
}

// validateArgs ensures at least one argument (the subcommand) was provided.
// If not, it prints usage and exits with code 1.
func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

// printChain iterates the blockchain from tip to genesis and prints each
// block's prev hash, transaction count, and hash. It also re-verifies proof-of-work
// by building work.BlockData from the block and calling Validate.
func (cli *CLI) printChain() {
	bci := cli.bc.Iterator()

	for {
		block := bci.Next()

		if block == nil {
			break
		}

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Transactions: %d\n", len(block.Transactions))
		fmt.Printf("Hash: %x\n", block.Hash)
		// Build BlockData so we can validate PoW without work importing block.
		pow := work.NewProofOfWork(&work.BlockData{
			PrevBlockHash:    block.PrevBlockHash,
			TransactionsHash: block.HashTransactions(),
			Timestamp:        block.Timestamp,
			Nonce:            block.Nonce,
		})
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) createBlockchain(address string) {
	bc := block.CreateBlockchain(address)
	bc.Close()
	fmt.Println("Done!")
}

// Run parses the first argument as the subcommand (addblock, printchain, etc.),
// then parses subcommand-specific flags and runs the appropriate handler.
// Exits with code 1 if the subcommand is unknown or required flags are missing.
func (cli *CLI) Run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block award to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")

	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}
}

// getBalance uses the blockchain already opened by main, finds all UTXOs for
// the given address, sums their values, and prints the balance.
func (cli *CLI) getBalance(address string) {
	balance := 0
	UTXOs := cli.bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

// send creates a new UTXO transaction from 'from' to 'to' for the given amount,
// mines a block containing that transaction, and appends it to the chain.
func (cli *CLI) send(from, to string, amount int) {
	tx := T.NewUTXOTransaction(from, to, amount, cli.bc)
	cli.bc.MineBlock([]*T.Transaction{tx})
	fmt.Println("Success!")
}
