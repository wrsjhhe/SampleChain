package cli

import (
	"fmt"
	"block"
)

func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := block.GetWallets(nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
}