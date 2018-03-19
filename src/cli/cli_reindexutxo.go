package cli

import (
	"block"
	"fmt"
)

func (cli *CLI) reindexUTXO(nodeID string) {
	bc := block.GetBlockchain(nodeID)
	UTXOSet := block.UTXOSet{bc}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}
