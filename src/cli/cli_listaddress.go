package cli

import (
	"fmt"
	"log"
	"block"
)

func (cli *CLI) listAddresses() {
	wallets, err := block.NewWallets()
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}
