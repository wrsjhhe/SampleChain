package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"block"
)

//CLI是为了执行命令行参数
type CLI struct {
	BC *block.Blockchain
}

func (cli *CLI)printUsage()  {
	fmt.Println("Usage:")
	fmt.Println(" addblock -data BLOCK_DATA - add a block to the blockchain")
	fmt.Println(" printchain - print all the blocks of the blockchain")
}

func (cli *CLI)validataArgs() {
	if len(os.Args)<2{
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI)addBlock(data string)  {
	cli.BC.AddBlock(data)
	fmt.Println("Success!")
}

func (cli *CLI) printChain()  {
	var bci = cli.BC.Iterator()

	for{
		var bc = bci.Next()

		fmt.Printf("Prev. hash: %x\n",bc.PrevBlockHash)
		fmt.Printf("Data: %s\n",bc.Data)
		fmt.Printf("Hash: %x\n",bc.Hash)
		var pow = block.NewProofOfWork(bc)
		fmt.Printf("Pow: %s\n",strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(bc.PrevBlockHash) == 0{
			break
		}
	}
}

func (cli *CLI)Run()  {
	cli.validataArgs()

	var addBlockCmd = flag.NewFlagSet("addblock",flag.ExitOnError)
	var printChainCmd = flag.NewFlagSet("printchian",flag.ExitOnError)

	var addBlockData = addBlockCmd.String("data","","Block data")



	switch os.Args[1] {
	case "addblock":
		var err = addBlockCmd.Parse(os.Args[2:])
		if err !=nil{
			log.Panic(err)
		}
	case "printchain":
		var err = printChainCmd.Parse(os.Args[2:])
		if err !=nil{
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)

	}

	if addBlockCmd.Parsed(){
		if *addBlockData == ""{
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed(){
		cli.printChain()
	}
}








