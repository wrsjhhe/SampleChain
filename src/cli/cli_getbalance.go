package cli

import (
	"fmt"
	"log"
	"block"
	"utils"
)

func (cli *CLI) getBalance(address, nodeID string) {
	if !block.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := block.GetBlockchain(nodeID)
	UTXOSet := block.UTXOSet{bc}
	defer bc.Db.Close()

	balance := 0
	pubKeyHash :=utils.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}