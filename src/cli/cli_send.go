package cli

import (
	"fmt"
	"log"
	"block"
)

func (cli *CLI) send(from, to string, amount int) {
	if !block.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !block.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := block.GetBlockchain()
	defer bc.Db.Close()

	tx := block.NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*block.Transaction{tx})
	fmt.Println("Success!")
}
