package cli

import (
	"fmt"
	"log"
	"block"
)

func (cli *CLI) send(from, to string, amount int, nodeID string, mineNow bool) {
	if !block.ValidateAddress(from) {
		log.Panic("ERROR: Sender address is not valid")
	}
	if !block.ValidateAddress(to) {
		log.Panic("ERROR: Recipient address is not valid")
	}

	bc := block.GetBlockchain(nodeID)
	UTXOSet := block.UTXOSet{bc}
	defer bc.Db.Close()

	wallets, err := block.GetWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := block.NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	if mineNow {
		cbTx := block.NewCoinbaseTX(from, "")
		txs := []*block.Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		block.SendTx(block.KnownNodes[0], tx)
	}

	fmt.Println("Success!")
}
