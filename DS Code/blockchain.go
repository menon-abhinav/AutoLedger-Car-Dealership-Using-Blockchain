package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"math"
	"reflect"
	"time"
)

func NewBlock(transactions []Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), nil, transactions, prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	hash := pow.Run()
	block.Hash = hash[:]

	transactionTypes := make([]string, len(transactions))
	for i, tx := range transactions {
		transactionTypes[i] = reflect.TypeOf(tx).Elem().Name() // Gets the type name of the transaction
		println("Transaction types : ", transactionTypes[i])
	}
	block.Transaction_types = transactionTypes
	return block
}

var (
	maxNonce = math.MaxInt64
)

func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	// Serialize each transaction individually and collect the results
	serializedTxs := make([][]byte, len(b.Transactions))
	for i, tx := range b.Transactions {
		serializedTx := tx.Serialize()
		serializedTxs[i] = serializedTx
	}

	// Create a struct that holds all block data, including serialized transactions
	tempBlock := struct {
		Timestamp         int64
		Transaction_types []string
		Transactions      [][]byte
		PrevBlockHash     []byte
		Hash              []byte
		Nonce             int
	}{
		Timestamp:         b.Timestamp,
		Transaction_types: b.Transaction_types,
		Transactions:      serializedTxs,
		PrevBlockHash:     b.PrevBlockHash,
		Hash:              b.Hash,
		Nonce:             b.Nonce,
	}

	// Encode the temporary block structure
	if err := encoder.Encode(tempBlock); err != nil {
		return nil
	}

	return result.Bytes()
}

func DeserializeBlock(data []byte) (*Block, error) {
	var tempBlock struct {
		Timestamp         int64
		Transaction_types []string
		Transactions      [][]byte
		PrevBlockHash     []byte
		Hash              []byte
		Nonce             int
	}

	decoder := gob.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(&tempBlock); err != nil {
		return nil, fmt.Errorf("failed to decode block: %w", err)
	}

	// Deserialize each transaction based on its type
	transactions := make([]Transaction, len(tempBlock.Transactions))
	for i, txData := range tempBlock.Transactions {
		txType := tempBlock.Transaction_types[i]
		var tx Transaction
		//println("DECODING TYPE ", txType)
		switch txType {
		case "VehicleRegistration":
			tx = &VehicleRegistration{}
			break
		case "VehicleSale":
			tx = &VehicleSale{}
			break
		case "LoanContract":
			tx = &LoanContract{}
			break
		case "genesis":
			tx = &genesis{}
			break
		default:
			println("Unknown transaction")
		}

		if err := gob.NewDecoder(bytes.NewReader(txData)).Decode(tx); err != nil {
			println("Error while decoding transaction")
			return nil, fmt.Errorf("failed to decode transaction: %w", err)
		}

		transactions[i] = tx
	}

	// Reconstruct the full Block
	block := &Block{
		Timestamp:         tempBlock.Timestamp,
		Transaction_types: tempBlock.Transaction_types,
		Transactions:      transactions,
		PrevBlockHash:     tempBlock.PrevBlockHash,
		Hash:              tempBlock.Hash,
		Nonce:             tempBlock.Nonce,
	}

	return block, nil
}
