package main

import (
	"block"
	"cli"
)

func main()  {

	var bc = block.NewBlockchain()

	defer bc.Db.Close()

	var cli1 = cli.CLI{bc}
	cli1.Run()

}

