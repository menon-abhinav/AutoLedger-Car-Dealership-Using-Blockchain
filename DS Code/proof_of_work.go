package main

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"strconv"
)

const targetBits = 12

// ProofOfWork represents a proof-of-work
type ProofOfWork struct {
	block  *Block
	target *big.Int
}

// NewProofOfWork builds and returns a ProofOfWork
func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{b, target}

	return pow
}

func (b *Block) prepareData() []byte {
	var headers []byte
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	b_nonce := []byte(strconv.Itoa(b.Nonce))
	headers = append(headers, b.PrevBlockHash...)
	headers = append(headers, timestamp...)
	headers = append(headers, b_nonce...)
	for _, tx := range b.Transactions {
		txData := (tx).Serialize() // Dereference the pointer and call Serialize on the Transaction
		headers = append(headers, txData...)
	}

	return headers
}

// Run performs a proof-of-work
func (pow *ProofOfWork) Run() []byte {
	var hashInt big.Int
	var hash [32]byte
	pow.block.Nonce = 0

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Transactions)
	for pow.block.Nonce < maxNonce {
		data := pow.block.prepareData()

		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			pow.block.Nonce++
		}
	}
	fmt.Print("\n\n")

	return hash[:]
}

// Validate validates block's PoW
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.block.prepareData()
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
