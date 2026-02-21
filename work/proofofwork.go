// Package work implements proof-of-work (Hashcash-style) mining for blocks.
// It does not import the block package to avoid an import cycle: block uses
// work for mining, so work only operates on a minimal BlockData struct.
package work

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go-blockchain/utils"
	"math"
	"math/big"
)

// targetBits = 24 matches common tutorials and early Bitcoin-style difficulty:
// the hash must have 24 leading zero bits (~6 hex digits), so on average
// ~2^24 hashes per block — realistic demo mining without taking forever.
const targetBits = 24

// maxNonce is the upper bound for the nonce in the mining loop. We use MaxInt64
// so that in practice we never hit it; if the target were unreachable the loop
// would still terminate instead of wrapping.
var maxNonce = math.MaxInt64

// BlockData holds the fields needed to compute and verify proof-of-work. We use
// this type instead of *block.Block so that the work package does not depend
// on block, which would create a cycle (block -> work -> block).
type BlockData struct {
	PrevBlockHash []byte
	Data          []byte
	Timestamp     int64
	Nonce         int
}

// ProofOfWork holds the block data and the difficulty target. The target is a
// big integer; a hash is valid only if its value as an integer is less than
// the target (i.e. has enough leading zero bits).
type ProofOfWork struct {
	block  *BlockData
	target *big.Int
}

// NewProofOfWork builds the PoW target and returns a ProofOfWork ready to Run.
// The target is 1 << (256 - targetBits), so a 256-bit hash must be below that
// value — equivalently, the hash must have at least targetBits leading zero bits.
func NewProofOfWork(b *BlockData) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
	pow := &ProofOfWork{b, target}
	return pow
}

// prepareData serializes the block header plus nonce into a single byte slice
// that we hash. The order and format must match exactly when mining and when
// validating, or the re-computed hash would not match. We include targetBits
// so that changing difficulty invalidates old blocks.
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			utils.IntToHex(pow.block.Timestamp),
			utils.IntToHex(int64(targetBits)),
			utils.IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

// Run performs mining: it tries nonces starting from 0 until it finds one such
// that SHA-256(prepareData(nonce)) is below the target. It returns that nonce
// and the resulting hash so the caller can store them on the block.
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)

	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			break
		}
		nonce++
	}
	fmt.Printf("%x\n", hash)
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// Validate recomputes the hash using the block's stored nonce and checks that
// it is below the target. This allows anyone to verify that the block was
// actually mined without re-running the full search.
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
