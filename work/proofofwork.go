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

var maxNonce = math.MaxInt64

// struct created to avoid import cycle
type BlockData struct {
	PrevBlockHash []byte
	Data          []byte
	Timestamp     int64
	Nonce         int
}

type ProofOfWork struct {
	block  *BlockData // pointer to block
	target *big.Int   // pointer to target
}

func NewProofOfWork(b *BlockData) *ProofOfWork {
	target := big.NewInt(1)
	// 256 is used below as its the length of the SHA-256 hash in bits
	// sshifting left gives -> 0x10000000000000000000000000000000000000000000000000000000000
	// occupies 29 bytes in memory
	target.Lsh(target, uint(256-targetBits))
	pow := &ProofOfWork{b, target}
	return pow
}

// nonce is a counter from the Hashcash description
// this is a cryptographic term
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

func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int // integer representation of hash
	var hash [32]byte
	nonce := 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)

	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Printf("%x\n", hash)
	fmt.Print("\n\n")

	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
