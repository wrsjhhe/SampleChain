package cli

import (
	"fmt"
	"log"
	"block"
)


func (cli *CLI) createBlockchain(address, nodeID string) {
	if !block.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := block.CreateBlockchain(address, nodeID)
	defer bc.Db.Close()

	UTXOSet := block.UTXOSet{bc}
	UTXOSet.Reindex()

	fmt.Println("Done!")
}
