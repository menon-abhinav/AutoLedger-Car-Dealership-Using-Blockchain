package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

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
