package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"math"
	"math/big"
	"os"
	"reflect"
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

type TransactionWrapper struct {
	TypeName        string
	TransactionData []byte
}

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

const dbFile = "blockchain.db"
const blocksBucket = "blocks"

// Blockchain keeps a sequence of Blocks
type Blockchain struct {
	tip []byte
	db  *bolt.DB
}

type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

// AddBlock saves provided data as a block in the blockchain
func (bc *Blockchain) AddBlock(t []Transaction) {
	var lastHash []byte

	err := bc.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	newBlock := NewBlock(t, lastHash)

	err = bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.tip = newBlock.Hash

		return nil
	})
}

func (bc *Blockchain) Iterator() *BlockchainIterator {
	bci := &BlockchainIterator{bc.tip, bc.db}

	return bci
}

// Next returns next block starting from the tip
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		//println("INSIDE NEXT 1", encodedBlock)
		block, _ = DeserializeBlock(encodedBlock)
		//println("INSIDE NEXT 2", block)
		return nil
	})

	if err != nil {
		log.Panic(err)

	}

	i.currentHash = block.PrevBlockHash
	//println("INSIDE NEXT 3", block)

	return block
}

// NewBlockchain creates a new Blockchain with genesis Block
func NewBlockchain() *Blockchain {
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))

		if b == nil {
			fmt.Println("No existing blockchain found. Creating a new one...")
			genesis := NewGenesisBlock()

			b, err := tx.CreateBucket([]byte(blocksBucket))
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(genesis.Hash, genesis.Serialize())
			if err != nil {
				log.Panic(err)
			}

			err = b.Put([]byte("l"), genesis.Hash)
			if err != nil {
				log.Panic(err)
			}
			tip = genesis.Hash
		} else {
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

type CLI struct {
	bc *Blockchain
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  addblock -tx TYPE [OPTIONS] - add a block with specified transaction to the blockchain")
	fmt.Println("    Transaction types and options:")
	fmt.Println("      VehicleRegistration -vin VIN -owner OWNER -date DATE")
	fmt.Println("      VehicleSale -vin VIN -dealer DEALER -buyer BUYER -date DATE -price PRICE")
	fmt.Println("      LoanContract -vin VIN -borrower BORROWER -lender LENDER -amount AMOUNT -start START -end END")
	fmt.Println("  printchain - print all the blocks of the blockchain")
}

func (cli *CLI) addBlock(txType string, args []string) {
	switch txType {
	case "VehicleRegistration":
		cli.addVehicleRegistration(args)
	case "VehicleSale":
		cli.addVehicleSale(args)
	case "LoanContract":
		cli.addLoanContract(args)
	default:
		fmt.Println("Unsupported transaction type")
		cli.printUsage()
	}
}

func (cli *CLI) addVehicleRegistration(args []string) {
	layout := "2006-01-02"
	cmd := flag.NewFlagSet("VehicleRegistration", flag.ExitOnError)
	vin := cmd.String("vin", "", "Vehicle Identification Number")
	owner := cmd.String("owner", "", "Owner's name or identifier")
	date := cmd.String("date", "", "Registration date as UNIX timestamp")

	// Parse the provided arguments according to the defined flags
	err := cmd.Parse(args)
	if err != nil {
		log.Panic(err)
	}

	// Validate VIN
	if *vin == "" {
		fmt.Println("A valid VIN is required.")
		cmd.Usage()
		os.Exit(1)
	}

	// Validate owner
	if *owner == "" {
		fmt.Println("An owner's identifier is required.")
		cmd.Usage()
		os.Exit(1)
	}

	// Validate registration date
	date_validated, err := time.Parse(layout, *date)
	if err != nil {
		log.Panic("Invalid start date format. Use YYYY-MM-DD.")
	}

	// Create the VehicleRegistration transaction
	vr := &VehicleRegistration{VIN: *vin, Owner: []byte(*owner), RegistrationDate: date_validated.Unix()}

	// Add the transaction to the blockchain
	cli.bc.AddBlock([]Transaction{vr})

	fmt.Println("Vehicle registration transaction added successfully!")
}

func (cli *CLI) addVehicleSale(args []string) {
	layout := "2006-01-02"
	cmd := flag.NewFlagSet("VehicleSale", flag.ExitOnError)
	vin := cmd.String("vin", "", "Vehicle Identification Number")
	dealer := cmd.String("dealer", "", "Dealer's identifier")
	buyer := cmd.String("buyer", "", "Buyer's identifier")
	date := cmd.String("date", "", "Sale date as UNIX timestamp")
	price := cmd.Int("price", 0, "Sale price")

	err := cmd.Parse(args)
	if err != nil {
		log.Panic(err)
	}

	// Validate VIN
	if *vin == "" {
		fmt.Println("A valid VIN is required.")
		cmd.Usage()
		os.Exit(1)
	}

	// Validate dealer and buyer
	if *dealer == "" || *buyer == "" {
		fmt.Println("Both dealer and buyer identifiers are required.")
		cmd.Usage()
		os.Exit(1)
	}

	// Validate sale price
	if *price <= 0 {
		fmt.Println("Sale price must be greater than 0.")
		cmd.Usage()
		os.Exit(1)
	}

	date_validated, err := time.Parse(layout, *date)
	if err != nil {
		log.Panic("Invalid start date format. Use YYYY-MM-DD.")
	}

	currentOwner, err := cli.bc.FindLatestOwnerByVIN(*vin)
	if err != nil {
		fmt.Println("Error: Could not find the vehicle with the specified VIN.")
		os.Exit(1)
	}

	if string(currentOwner) != *dealer {
		fmt.Println("Error: The dealer is not the current owner of the vehicle.")
		os.Exit(1)
	}

	// 2. Check if the vehicle is under any active loan.
	if cli.hasActiveLoan(*vin) {
		fmt.Println("Error: The vehicle is currently under an active loan and cannot be sold.")
		os.Exit(1)
	}

	// Create the VehicleSale transaction
	vs := &VehicleSale{
		VIN:      *vin,
		Dealer:   []byte(*dealer),
		Buyer:    []byte(*buyer),
		SaleDate: date_validated.Unix(),
		Price:    *price,
	}

	// Add the transaction to the blockchain
	cli.bc.AddBlock([]Transaction{vs})

	fmt.Println("Vehicle sale transaction added successfully!")
}

func (cli *CLI) addLoanContract(args []string) {

	layout := "2006-01-02"
	cmd := flag.NewFlagSet("LoanContract", flag.ExitOnError)
	vin := cmd.String("vin", "", "Vehicle Identification Number (VIN)")
	borrower := cmd.String("borrower", "", "Borrower's identifier")
	lender := cmd.String("lender", "", "Lender's identifier")
	loanAmount := cmd.Int("amount", 0, "The amount of the loan")
	startDateStr := cmd.String("start", "", "The start date of the loan in YYYY-MM-DD format")
	endDateStr := cmd.String("end", "", "The end date of the loan in YYYY-MM-DD format")

	err := cmd.Parse(args)
	if err != nil {
		log.Panic(err)
	}

	// Validate VIN
	if *vin == "" {
		fmt.Println("A valid VIN is required.")
		cmd.Usage()
		os.Exit(1)
	}

	// Validate borrower and lender
	if *borrower == "" || *lender == "" {
		fmt.Println("Both borrower and lender identifiers are required.")
		cmd.Usage()
		os.Exit(1)
	}

	// Validate loan amount
	if *loanAmount <= 0 {
		fmt.Println("Loan amount must be greater than 0.")
		cmd.Usage()
		os.Exit(1)
	}

	startDate, err := time.Parse(layout, *startDateStr)
	if err != nil {
		log.Panic("Invalid start date format. Use YYYY-MM-DD.")
	}
	endDate, err := time.Parse(layout, *endDateStr)
	if err != nil {
		log.Panic("Invalid end date format. Use YYYY-MM-DD.")
	}

	currentOwner, err := cli.bc.FindLatestOwnerByVIN(*vin)
	if err != nil {
		fmt.Println("Error: Could not find the vehicle with the specified VIN.")
		os.Exit(1)
	}

	if string(currentOwner) != *borrower {
		fmt.Println("Error: The borrower is not the current owner of the vehicle.")
		os.Exit(1)
	}

	// Create the LoanContract transaction
	lc := &LoanContract{
		VIN:        *vin,
		Borrower:   []byte(*borrower),
		Lender:     []byte(*lender),
		LoanAmount: *loanAmount,
		StartDate:  startDate.Unix(),
		EndDate:    endDate.Unix(),
	}

	// Add the transaction to the blockchain
	cli.bc.AddBlock([]Transaction{lc})

	fmt.Println("Loan contract transaction added successfully!")
}
func (cli *CLI) printChain() {
	bci := cli.bc.Iterator()

	for {
		block := bci.Next()
		if block == nil {
			fmt.Println("Reached the end of the blockchain or encountered an error.")
			break
		}

		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)

		for _, tx := range block.Transactions {
			tx.print_transaction()
			fmt.Println()
		}
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := NewProofOfWork(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}

	// Check for the 'addblock' command specifically
	if os.Args[1] == "addblock" && len(os.Args) < 3 {
		fmt.Println("Error: 'addblock' command requires a transaction type.")
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {
	cli.validateArgs()
	println(os.Args[1])
	switch os.Args[1] {
	case "addblock":
		if len(os.Args) < 3 {
			cli.printUsage()
			os.Exit(1)
		}
		txType := os.Args[2]
		switch txType {
		case "VehicleRegistration":
			if len(os.Args) < 4 {
				cli.printUsage()
				os.Exit(1)
			}
			cli.addVehicleRegistration(os.Args[3:])
			break
		case "VehicleSale":
			if len(os.Args) < 4 {
				cli.printUsage()
				os.Exit(1)
			}
			cli.addVehicleSale(os.Args[3:])
			break
		case "LoanContract":
			if len(os.Args) < 4 {
				cli.printUsage()
				os.Exit(1)
			}
			cli.addLoanContract(os.Args[3:])
			break
		default:
			fmt.Println("Unsupported transaction type:", txType)
			cli.printUsage()
			os.Exit(1)
		}
		break
	case "printchain":
		//println("CALLING PRINTCHAIN")
		cli.printChain()
	default:
		cli.printUsage()
		os.Exit(1)
	}
}

func (bc *Blockchain) FindLatestOwnerByVIN(vin string) ([]byte, error) {
	bci := bc.Iterator()
	for {
		block := bci.Next()

		// Check transactions in reverse order since the latest transaction will be at the end
		for i := len(block.Transactions) - 1; i >= 0; i-- {
			tx := block.Transactions[i]
			switch tx := tx.(type) {
			case *VehicleRegistration:
				if tx.VIN == vin {
					return tx.Owner, nil
				}
			case *VehicleSale:
				if tx.VIN == vin {
					return tx.Buyer, nil
				}
			}
		}

		// Genesis block reached
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
	return nil, errors.New("vehicle not found")
}

func (cli *CLI) hasActiveLoan(vin string) bool {
	bci := cli.bc.Iterator()
	currentTime := time.Now().Unix()

	for {
		block := bci.Next()

		// Iterate through transactions in the block
		for _, tx := range block.Transactions {
			switch tx := tx.(type) {
			case *LoanContract:
				if tx.VIN == vin {
					// Check if the loan end date is in the future, indicating an active loan
					if tx.EndDate > currentTime {
						return true
					}
				}
			}
		}

		// Stop iteration if we've reached the genesis block
		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return false
}

func main() {

	gob.Register(&VehicleRegistration{})
	gob.Register(&VehicleSale{})
	gob.Register(&LoanContract{})
	gob.Register(&genesis{})
	gob.Register(&Block{})
	bc := NewBlockchain()
	defer bc.db.Close()

	cli := CLI{bc}
	cli.Run()
}
