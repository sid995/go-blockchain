package block

import (
	"go-blockchain/utils"
	"log"

	"github.com/boltdb/bolt"
)

const dbFile = "blockchain.db"       // path to the BoltDB file
const blocksBucket = "blocks"        // bucket name for block hash -> serialized block

// Blockchain persists blocks in BoltDB. We store only the chain tip in memory;
// full blocks are read from the DB when needed (e.g. via the iterator).
type Blockchain struct {
	tip []byte   // hash of the latest block; key "l" in the bucket also stores this
	db  *bolt.DB
}

// BlockchainIterator walks the chain from tip back to genesis. currentHash is
// the hash of the block to return on the next call to Next(); nil means end of chain.
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// AddBlock reads the current tip from the DB, mines a new block linked to it,
// then persists the new block and updates the tip key ("l") in BoltDB.
func (bc *Blockchain) AddBlock(data string) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(data, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		if err := b.Put(newBlock.Hash, newBlock.SerializeBlock()); err != nil {
			return err
		}
		if err = b.Put([]byte("l"), newBlock.Hash); err != nil {
			return err
		}
		bc.tip = newBlock.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

}

// NewGenesisBlock creates the first block in the chain. It has no previous
// block (empty PrevBlockHash). The content "Genesis Block" is conventional;
// in Bitcoin the genesis block has fixed, hardcoded data.
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

// NewBlockchain opens or creates the DB file. If the blocks bucket is empty,
// it creates the genesis block and stores it; otherwise it loads the existing
// tip. So the chain always has at least one block for AddBlock to link to.
func NewBlockchain() *Blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			genesis := NewGenesisBlock()
			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				return err
			}

			if err = b.Put(genesis.Hash, genesis.SerializeBlock()); err != nil {
				return err
			}
			if err = b.Put([]byte("l"), genesis.Hash); err != nil {
				return err
			}
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}
		return nil
	})

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
