package block

// Blockchain holds an ordered list of blocks. In this demo we use an in-memory
// slice; a real implementation would persist to disk and might use a different
// structure (e.g. index by hash) for lookups.
type Blockchain struct {
	blocks []*Block
}

// Blocks returns a slice of all blocks in the chain. Callers can iterate but
// should not modify the slice; we expose it for simplicity in this demo.
func (bc *Blockchain) Blocks() []*Block {
	return bc.blocks
}

// AddBlock creates a new block with the given data, links it to the current tip
// (last block), runs proof-of-work, and appends it to the chain. The previous
// block's hash is used so the chain is cryptographically linked.
func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	newBlock := NewBlock(data, prevBlock.Hash)
	bc.blocks = append(bc.blocks, newBlock)
}

// NewGenesisBlock creates the first block in the chain. It has no previous
// block (empty PrevBlockHash). The content "Genesis Block" is conventional;
// in Bitcoin the genesis block has fixed, hardcoded data.
func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block", []byte{})
}

// NewBlockchain returns a new blockchain containing only the genesis block.
// We need at least one block so that AddBlock always has a previous block to
// link to when computing PrevBlockHash.
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{NewGenesisBlock()}}
}
