package cli

import (
	"block"
	"fmt"
)

func (cli *CLI)reindexUTXO()  {
	var bc = block.GetBlockchain()
	var utxoSet = block.UTXOSet{bc}
	utxoSet.Reindex()

	var count = utxoSet.CountTransactions()
	fmt.Printf("Done!There are %d transaction in the UTXO set.\n",count)
}
