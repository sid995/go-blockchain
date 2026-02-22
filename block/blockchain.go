package block

import (
	"encoding/hex"
	"fmt"
	"go-blockchain/utils"
	"log"
	"os"

	Tx "go-blockchain/transaction"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db" // path to the BoltDB file
const blocksBucket = "blocks"  // bucket name for block hash -> serialized block
// genesisCoinbaseData is the message in the first block's coinbase (Bitcoin reference).
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"

// TxOutputs is a slice of transaction outputs; alias to avoid repeating []Tx.TxOutput.
type TxOutputs []Tx.TxOutput

// Blockchain persists blocks in BoltDB. We store only the chain tip in memory;
// full blocks are read from the DB when needed (e.g. via the iterator).
type Blockchain struct {
	tip []byte // hash of the latest block; key "l" in the bucket also stores this
	db  *bolt.DB
}

// BlockchainIterator walks the chain from tip back to genesis. currentHash is
// the hash of the block to return on the next call to Next(); nil means end of chain.
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// CreateBlockchain creates a new blockchain DB and genesis block, paying the
// initial coinbase reward to address. Fails if the DB file already exists.
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain already exists")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbtx := Tx.NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := NewGenesisBlock(cbtx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.SerializeBlock())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

// NewBlockchain opens or creates the DB file. If the blocks bucket is empty,
// it creates the genesis block and stores it; otherwise it loads the existing
// tip. So the chain always has at least one block for AddBlock to link to.
func NewBlockchain(address string) *Blockchain {
	if dbExists() == false {
		fmt.Println("No existing blockchain found. Create one first.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

// Iterator returns an iterator that walks from the current tip back to genesis.
// Safe to call multiple times; each call starts from the tip.
func (bc *Blockchain) Iterator() *BlockchainIterator {
	return &BlockchainIterator{bc.tip, bc.db}
}

// Next returns the block for the current hash and advances the iterator to the
// previous block. Returns nil when currentHash is nil (before first call or
// after returning genesis), or when the block cannot be read from the DB.
// End-of-chain: after returning the genesis block we set currentHash to nil
// via SliceOrNil(PrevBlockHash) so the next Next() returns nil.
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	if i.currentHash == nil {
		return nil
	}

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		if encodedBlock == nil {
			return nil
		}
		block = DeserializeBlock(encodedBlock)

		return nil
	})

	if err != nil {
		return nil
	}
	if block == nil {
		return nil
	}

	// Advance to previous block; genesis has empty PrevBlockHash, so SliceOrNil
	// yields nil and the next Next() will return nil.
	i.currentHash = utils.SliceOrNil(block.PrevBlockHash)
	return block
}

// Close releases the blockchain's DB connection. Call when done (e.g. defer from main).
func (bc *Blockchain) Close() error {
	return bc.db.Close()
}

// dbExists returns true if the blockchain DB file already exists on disk.
func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

// FindUnspentTransactions returns all transactions that contain at least one
// output spendable by address and that have not been spent by any later input.
// We iterate the chain and track which outputs are consumed by inputs that
// address can unlock.
func (bc *Blockchain) FindUnspentTransactions(address string) []Tx.Transaction {
	var unspentTXs []Tx.Transaction
	spentTXOs := make(map[string][]int) // txID -> slice of output indices already spent
	bci := bc.Iterator()

	for {
		block := bci.Next()
		if block == nil {
			break
		}

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// Skip if this output was already spent in a later block.
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

// FindUTXO returns all unspent transaction outputs that can be unlocked by
// address. It uses FindUnspentTransactions and then collects the matching
// outputs from each of those transactions.
func (bc *Blockchain) FindUTXO(address string) TxOutputs {
	var UTXOs TxOutputs
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

// FindSpendableOutputs finds enough unspent outputs for address to cover amount.
// Returns the total value accumulated and a map of txID -> output indices to use
// as inputs. Used by NewUTXOTransaction to build the input list.
func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

// MineBlock creates a new block with the given transactions, mines it (PoW),
// and appends it to the chain. Same as AddBlock but named for CLI clarity.
func (bc *Blockchain) MineBlock(transactions []*Tx.Transaction) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(transactions, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.SerializeBlock())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}
}
