package cli

import (
	"fmt"
	"strconv"
	"block"
)

func (cli *CLI) printChain(nodeID string) {
	bc := block.GetBlockchain(nodeID)
	defer bc.Db.Close()

	bci := bc.Iterator()

	for {
		blck := bci.Next()

		fmt.Printf("============ Block %x ============\n", blck.Hash)
		fmt.Printf("Height: %d\n", blck.Height)
		fmt.Printf("Prev. block: %x\n", blck.PrevBlockHash)
		pow := block.NewProofOfWork(blck)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range blck.Transactions {
			fmt.Println(tx)
		}
		fmt.Printf("\n\n")

		if len(blck.PrevBlockHash) == 0 {
			break
		}
	}
}