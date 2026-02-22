// Package block defines the block structure and blockchain creation.
// Blocks are created via proof-of-work mining; this package depends on work
// for hashing, but work must not depend on block (to avoid import cycles).
package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	Tx "go-blockchain/transaction"
	"go-blockchain/work"
	"log"
	"strconv"
	"time"
)

// Block is a single link in the blockchain. It holds the data and metadata
// needed to chain blocks together and verify integrity.
//
// In a full blockchain this would also include version, merkle root of
// transactions, etc. We keep it minimal: previous block hash (for chaining),
// payload data, timestamp, and the PoW result (hash + nonce).
type Block struct {
	Timestamp     int64 // Unix time when the block was created; used in PoW input
	Transactions  []*Tx.Transaction
	PrevBlockHash []byte // Hash of the previous block; genesis block has nil/empty
	Hash          []byte // SHA-256 hash meeting the PoW target; set after mining
	Nonce         int    // Nonce that produced a valid hash; needed to re-verify PoW later
}

// SetHash computes a hash of the block header (PrevBlockHash, Hash, timestamp)
// and stores it in b.Hash. In this codebase blocks are created via NewBlock
// with proof-of-work, so SetHash is optional; kept for compatibility or tests.
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Hash, timestamp}, []byte{})
	hash := sha256.Sum256(headers)

	b.Hash = hash[:]
}

// NewBlock creates a new block with the given data and previous block hash,
// then runs proof-of-work to find a nonce such that the block hash meets the
// difficulty target. The block's Hash and Nonce are set from the mining result.
//
// We pass work.BlockData (not *Block) to work.NewProofOfWork so that the work
// package does not need to import block, avoiding a circular dependency.
func NewBlock(transactions []*Tx.Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}

	workBlockData := &work.BlockData{
		PrevBlockHash:    block.PrevBlockHash,
		TransactionsHash: hashTransactionList(block.Transactions),
		Timestamp:        block.Timestamp,
		Nonce:            block.Nonce,
	}

	pow := work.NewProofOfWork(workBlockData)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// NewGenesisBlock creates the first block in the chain. It has no previous
// block (empty PrevBlockHash). The content "Genesis Block" is conventional;
// in Bitcoin the genesis block has fixed, hardcoded data.
func NewGenesisBlock(coinbase *Tx.Transaction) *Block {
	return NewBlock([]*Tx.Transaction{coinbase}, []byte{})
}

// HashTransactions returns a single hash representing all transactions in the
// block (concatenate tx IDs, then SHA-256). Used in PoW and for verification.
func (b *Block) HashTransactions() []byte {
	return hashTransactionList(b.Transactions)
}

// hashTransactionList hashes the list of transactions the same way HashTransactions
// does, so the block package can pass this hash to work.BlockData without work
// needing the full transaction list or importing block.
func hashTransactionList(transactions []*Tx.Transaction) []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range transactions {
		txHashes = append(txHashes, tx.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

// SerializeBlock encodes the block to bytes using gob so it can be stored in
// BoltDB. The same encoding is used when reading back with DeserializeBlock.
func (b *Block) SerializeBlock() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	if err := encoder.Encode(b); err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock decodes a gob-encoded block from d. Callers must pass
// non-nil data (e.g. the iterator checks encodedBlock == nil before calling).
func DeserializeBlock(d []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(d))
	if err := decoder.Decode(&block); err != nil {
		log.Panic(err)
	}
	return &block
}
