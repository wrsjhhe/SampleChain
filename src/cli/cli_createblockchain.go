package cli

import (
	"fmt"
	"log"
	"block"
)

func (cli *CLI) createBlockchain(address string) {
	if ! block.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := block.CreateBlockchain(address)
	bc.Db.Close()
	fmt.Println("Done!")
}
