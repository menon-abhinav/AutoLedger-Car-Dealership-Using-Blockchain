package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"math/big"
	"strconv"
	"time"
)

type VehicleRegistration struct {
	VIN              string
	Owner            []byte
	RegistrationDate int64
}

type VehicleSale struct {
	VIN      string
	Dealer   []byte
	Buyer    []byte
	SaleDate int64
	Price    int
}

type LoanContract struct {
	VIN        string
	Borrower   []byte
	Lender     []byte
	LoanAmount int
	StartDate  int64
	EndDate    int64
}

type genesis struct {
	VIN string
	//data string
}

type Block struct {
	Timestamp     int64
	Transactions  []Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

type Transaction interface {
	ID() string        // Unique identifier for the transaction, typically the VIN for vehicle-related transactions.
	Serialize() []byte // Converts the transaction data into a byte slice for hashing.
}

func (vr *genesis) ID() string {
	return vr.VIN
}

func (vr *genesis) Serialize() []byte {
	type Wrapper struct {
		Type    string
		genesis *genesis
	}

	wrapped := Wrapper{
		Type:    "genesis",
		genesis: vr,
	}

	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	if err := encoder.Encode(wrapped); err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func (vr *VehicleRegistration) ID() string {
	return vr.VIN
}

func (vr *VehicleRegistration) Serialize() []byte {
	type Wrapper struct {
		Type                string
		VehicleRegistration *VehicleRegistration
	}

	wrapped := Wrapper{
		Type:                "VehicleRegistration",
		VehicleRegistration: vr,
	}

	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	if err := encoder.Encode(wrapped); err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func (vr *VehicleSale) ID() string {
	return vr.VIN
}

func (vr *VehicleSale) Serialize() []byte {
	type Wrapper struct {
		Type        string
		VehicleSale *VehicleSale
	}

	wrapped := Wrapper{
		Type:        "VehicleSale",
		VehicleSale: vr,
	}

	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	if err := encoder.Encode(wrapped); err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func (vr *LoanContract) ID() string {
	return vr.VIN
}

func (vr *LoanContract) Serialize() []byte {
	type Wrapper struct {
		Type         string
		LoanContract *LoanContract
	}

	wrapped := Wrapper{
		Type:         "LoanContract",
		LoanContract: vr,
	}

	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	if err := encoder.Encode(wrapped); err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func NewGenesisBlock() *Block {
	gen_trans := &genesis{"GENESIS BLOCK"}
	newBlock := NewBlock([]Transaction{gen_trans}, []byte{})
	return newBlock
}

func DeserializeTransaction(data []byte) (Transaction, error) {
	var wrapper struct {
		Type                string
		VehicleRegistration *VehicleRegistration
		VehicleSale         *VehicleSale
		LoanContract        *LoanContract
		genesis             *genesis
	}

	gobReader := bytes.NewReader(data)
	gobDecoder := gob.NewDecoder(gobReader)

	if err := gobDecoder.Decode(&wrapper); err != nil {
		log.Panic(err)
	}

	switch wrapper.Type {
	case "VehicleRegistration":
		return wrapper.VehicleRegistration, nil
	case "VehicleSale":
		return wrapper.VehicleSale, nil
	case "LoanContract":
		return wrapper.LoanContract, nil
	case "genesis":
		return wrapper.genesis, nil
	default:
		return nil, fmt.Errorf("unknown transaction type: %s", wrapper.Type)
	}
}

//func (b *Block) SetHash() {
//	var headers []byte
//	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
//	nonce := []byte(strconv.Itoa(b.Nonce))
//	headers = append(headers, b.PrevBlockHash...)
//	headers = append(headers, timestamp...)
//	headers = append(headers, nonce...)
//	for _, tx := range b.Transactions {
//		txData := (tx).Serialize() // Dereference the pointer and call Serialize on the Transaction
//		headers = append(headers, txData...)
//	}
//
//	hash := sha256.Sum256(headers)
//	b.Hash = hash[:]
//}

func NewBlock(transactions []Transaction, prevBlockHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevBlockHash, []byte{}, 0}
	pow := NewProofOfWork(block)
	hash := pow.Run()
	block.Hash = hash[:]
	return block
}

var (
	maxNonce = math.MaxInt64
)

const targetBits = 20

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

func main() {
	// Initialize the blockchain with the genesis block
	genesisBlock := NewGenesisBlock()
	blockchain := []*Block{genesisBlock}

	// Transaction data
	vr := &VehicleRegistration{
		VIN:              "1HGCM82633A004352",
		Owner:            []byte("John Doe"),
		RegistrationDate: time.Now().Unix(),
	}
	vs := &VehicleSale{
		VIN:      "2FMDK39C07BBB8567",
		Dealer:   []byte("Car Dealer"),
		Buyer:    []byte("Jane Doe"),
		SaleDate: time.Now().Unix(),
		Price:    15000,
	}
	lc := &LoanContract{
		VIN:        "JH4KA7660MC000000",
		Borrower:   []byte("Jake Doe"),
		Lender:     []byte("Bank"),
		LoanAmount: 10000,
		StartDate:  time.Now().Unix(),
		EndDate:    time.Now().AddDate(1, 0, 0).Unix(),
	}

	newBlock := NewBlock([]Transaction{vr, vs, lc}, blockchain[len(blockchain)-1].Hash)

	blockchain = append(blockchain, newBlock)
	// Print the blockchain
	for i, block := range blockchain {
		fmt.Printf("Block %d:\n", i)
		fmt.Printf("Timestamp: %d\n", block.Timestamp)
		fmt.Printf("Prev. Hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		for _, tx := range block.Transactions {
			fmt.Printf("Transaction: %+v\n", tx)
		}
		fmt.Println()
	}
}
