package cli

import (
	"fmt"
	"strconv"
	"block"
)

func (cli *CLI) printChain() {
	bc := block.GetBlockchain()
	defer bc.Db.Close()

	bci := bc.Iterator()

	for {
		nbc := bci.Next()

		fmt.Printf("============ Block %x ============\n", nbc.Hash)
		fmt.Printf("Prev. block: %x\n", nbc.PrevBlockHash)
		pow := block.NewProofOfWork(nbc)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range nbc.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(nbc.PrevBlockHash) == 0 {
			break
		}
	}
}
