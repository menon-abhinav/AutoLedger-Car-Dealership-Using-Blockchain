package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
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
	Timestamp         int64
	Transaction_types []string
	Transactions      []Transaction
	PrevBlockHash     []byte
	Hash              []byte
	Nonce             int
}

type Transaction interface {
	ID() string        // Unique identifier for the transaction, typically the VIN for vehicle-related transactions.
	Serialize() []byte // Converts the transaction data into a byte slice for hashing.
	print_transaction()
}

func (vr *genesis) ID() string {
	return vr.VIN
}

func (vr *genesis) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	if err := encoder.Encode(vr); err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func (vr *genesis) print_transaction() {
	fmt.Println("Genesis Block")
	fmt.Printf("ID: %s\n", vr.VIN)
}

func (vr *VehicleRegistration) ID() string {
	return vr.VIN
}

func (vr *VehicleRegistration) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	if err := encoder.Encode(vr); err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func (vr *VehicleRegistration) print_transaction() {
	fmt.Println("Vehicle Registration Transaction")
	fmt.Printf("ID: %s\n", vr.VIN)
	fmt.Printf("Owner: %s\n", string(vr.Owner))                                                   // Convert byte slice to string
	fmt.Printf("Registration Date: %s\n", time.Unix(vr.RegistrationDate, 0).Format("2006-01-02")) // Format Unix timestamp
}

func (vr *VehicleSale) ID() string {
	return vr.VIN
}

func (vr *VehicleSale) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	if err := encoder.Encode(vr); err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func (vr *VehicleSale) print_transaction() {
	fmt.Println("Vehicle Sale Transaction")
	fmt.Printf("ID: %s\n", vr.VIN)
	fmt.Printf("Dealer: %s\n", string(vr.Dealer)) // Convert byte slice to string
	fmt.Printf("Buyer: %s\n", string(vr.Buyer))   // Convert byte slice to string
	fmt.Printf("Price: %d\n", vr.Price)
	fmt.Printf("Sale Date: %s\n", time.Unix(vr.SaleDate, 0).Format("2006-01-02")) // Format Unix timestamp
}

func (vr *LoanContract) ID() string {
	return vr.VIN
}

func (vr *LoanContract) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	if err := encoder.Encode(vr); err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func (vr *LoanContract) print_transaction() {
	fmt.Println("Loan Contract Transaction")
	fmt.Printf("ID: %s\n", vr.VIN)
	fmt.Printf("Lender: %s\n", string(vr.Lender))     // Convert byte slice to string
	fmt.Printf("Borrower: %s\n", string(vr.Borrower)) // Convert byte slice to string
	fmt.Printf("Loan Amount: %d\n", vr.LoanAmount)
	fmt.Printf("Start Date: %s\n", time.Unix(vr.StartDate, 0).Format("2006-01-02")) // Format Unix timestamp
	fmt.Printf("End Date: %s\n", time.Unix(vr.EndDate, 0).Format("2006-01-02"))     // Format Unix timestamp
}

func NewGenesisBlock() *Block {
	gen_trans := &genesis{"GENESIS BLOCK"}
	newBlock := NewBlock([]Transaction{gen_trans}, []byte{})
	return newBlock
}
