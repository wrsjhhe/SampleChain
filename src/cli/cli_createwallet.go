package cli

import (
	"fmt"
	"block"
)

func (cli *CLI) createWallet() {
	wallets, _ := block.GetWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}