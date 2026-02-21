// Package block defines the block structure and blockchain creation.
// Blocks are created via proof-of-work mining; this package depends on work
// for hashing, but work must not depend on block (to avoid import cycles).
package block

import (
	"bytes"
	"crypto/sha256"
	"go-blockchain/work"
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
	Timestamp     int64  // Unix time when the block was created; used in PoW input
	Data          []byte // Payload (e.g. transaction data); in Bitcoin this would be replaced by a merkle root
	PrevBlockHash []byte // Hash of the previous block; genesis block has nil/empty
	Hash          []byte // SHA-256 hash meeting the PoW target; set after mining
	Nonce         int    // Nonce that produced a valid hash; needed to re-verify PoW later
}

// SetHash computes a simple hash of the block header (prev hash + data + timestamp)
// and stores it in b.Hash. We use this for a naive hash; in this codebase blocks
// are actually created via NewBlock which uses proof-of-work instead. SetHash is
// kept for compatibility or alternative use (e.g. blocks that don't use PoW).
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)

	b.Hash = hash[:]
}

// NewBlock creates a new block with the given data and previous block hash,
// then runs proof-of-work to find a nonce such that the block hash meets the
// difficulty target. The block's Hash and Nonce are set from the mining result.
//
// We pass work.BlockData (not *Block) to work.NewProofOfWork so that the work
// package does not need to import block, avoiding a circular dependency.
func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevBlockHash, []byte{}, 0}
	pow := work.NewProofOfWork(&work.BlockData{
		PrevBlockHash: block.PrevBlockHash,
		Data:          block.Data,
		Timestamp:     block.Timestamp,
		Nonce:         block.Nonce,
	})
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}
