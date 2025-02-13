package main

import (
	"encoding/gob"
)

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
