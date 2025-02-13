package main

import (
	"errors"
	"time"
)

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
