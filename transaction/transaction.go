// Package transaction defines the UTXO-style transaction model: inputs reference
// previous outputs, outputs have value and a lock (ScriptPubKey). The package
// does not import block; it uses SpendableOutputFinder to avoid an import cycle.
package transaction

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

// subsidy is the reward for mining a block (coinbase output value).
const subsidy = 10

// SpendableOutputFinder is implemented by *block.Blockchain. We use an interface
// here so the transaction package does not import block (which would create
// block -> transaction -> block). Callers pass the blockchain when creating
// a new UTXO transaction.
type SpendableOutputFinder interface {
	FindSpendableOutputs(address string, amount int) (int, map[string][]int)
}

// Transaction represents a single transfer: a list of inputs (referencing prior
// outputs) and a list of outputs (value + lock). ID is the hash of the tx.
type Transaction struct {
	ID   []byte
	Vin  []TxInput
	Vout []TxOutput
}

// SetID serializes the transaction with gob and sets ID to the SHA-256 hash of
// that encoding. Must be called after constructing a tx before using its ID.
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

// IsCoinbase returns true if this transaction has a single input with empty Txid
// and Vout == -1, i.e. it is the block reward transaction (no prior output).
func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

// TxOutput represents an output: an amount (Value) and a lock (ScriptPubKey).
// In a full implementation ScriptPubKey would be a script; here we use the
// address string for simplicity. Only an input with matching ScriptSig can spend it.
type TxOutput struct {
	Value        int
	ScriptPubKey string
}

// TxInput references a previous output by transaction ID and output index (Vout).
// ScriptSig is the unlock data; it must match the referenced output's ScriptPubKey.
type TxInput struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}

// NewCoinbaseTX creates the special "coinbase" transaction for block rewards.
// It has one input (empty txid, Vout -1) and one output paying subsidy to 'to'.
// The optional data string is stored in the input (e.g. block height or message).
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txin := TxInput{[]byte{}, -1, data}
	txout := TxOutput{subsidy, to}
	tx := Transaction{nil, []TxInput{txin}, []TxOutput{txout}}
	tx.SetID()

	return &tx
}

// CanUnlockOutputWith returns true if this input's ScriptSig matches the given
// unlocking data (e.g. the address that owns the referenced output).
func (in *TxInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

// CanBeUnlockedWith returns true if the given unlocking data matches this
// output's ScriptPubKey (i.e. the output can be spent by that key/address).
func (out *TxOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

// NewUTXOTransaction builds a transaction that spends from 'from's UTXOs to pay
// 'amount' to 'to', returning any change to 'from'. It uses chain to find
// spendable outputs; chain is typically *block.Blockchain. Panics if from has
// insufficient funds.
func NewUTXOTransaction(from, to string, amount int, chain SpendableOutputFinder) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	acc, validOutputs := chain.FindSpendableOutputs(from, amount)
	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TxInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, TxOutput{amount, to})

	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, from}) // change back to sender
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}
