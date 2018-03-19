package cli

import (
	"fmt"
	"log"
	"block"
)

func (cli *CLI) listAddresses(nodeID string) {
	wallets, err := block.GetWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}
