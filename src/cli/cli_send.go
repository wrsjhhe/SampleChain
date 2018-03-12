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

	var bc = block.GetBlockchain()
	var utxoSet = block.UTXOSet{bc}
	defer bc.Db.Close()

	var tx = block.NewUTXOTransaction(from, to, amount, bc)
	var cbTx = block.NewCoinbaseTX(from,"")
	var txs = []*block.Transaction{cbTx,tx}

	var newBlock = bc.MineBlock(txs)
	utxoSet.Update(newBlock)

	fmt.Println("Success!")
}
