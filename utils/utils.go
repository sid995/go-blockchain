// Package utils provides small helpers used across the blockchain packages:
// IntToHex for deterministic integer encoding in hashes, ReverseBytes for
// byte-order conversions, SliceOrNil for iterator end-of-chain. Kept in a
// separate package so block and work can both use it without depending on each other.
package utils

import (
	"bytes"
	"encoding/binary"
	"log"
)

// IntToHex converts an int64 to a fixed-size big-endian byte slice. We use
// big-endian so that byte order is consistent and hashes are deterministic;
// the proof-of-work and block packages use this for timestamp, targetBits, and
// nonce when building the data that gets hashed.
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// ReverseBytes reverses the given byte slice in place. Useful when working
// with little-endian vs big-endian representations (e.g. some Bitcoin-style
// encodings). Not used in the current PoW path but kept for compatibility
// or future use.
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

// SliceOrNil returns nil if s is nil or has length zero, otherwise returns s.
// Used so the blockchain iterator can represent "no next block" with nil
// (e.g. after genesis, PrevBlockHash is empty and we set currentHash = nil).
func SliceOrNil(s []byte) []byte {
	if len(s) == 0 {
		return nil
	}
	return s
}
