package block

import (
	"bytes"
	"crypto/sha256"
	"go-blockchain/work"
	"strconv"
	"time"
)

// Block struct created as part of blockchain
// Store valuable information
// Can store
//   - transactions
//   - version
//   - timestamp
//   - hash
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

// Calculating hash value for block
func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{b.PrevBlockHash, b.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)

	b.Hash = hash[:]
}

// Create new block with expected values and hashes
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
