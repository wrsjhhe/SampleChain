package cli

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"block"
	"utils"
)

//CLI是为了执行命令行参数
type CLI struct {}

func (cli *CLI)createBlockchain(address string)  {
	var bc = block.CreateBlockchain(address)
	bc.Db.Close()
	fmt.Println("Done")
}

func (cli *CLI)getBalance(address string)  {
	var bc = block.GetBlockchain(address)
	defer bc.Db.Close()

	var balance = 0
	var UTXOs = bc.FindUTXO(address)

	for _,out:=range UTXOs{
		balance += out.Value
	}

	fmt.Printf("Balance of '%s' : %d\n",address,balance)
}


//打印帮助
func (cli *CLI)printUsage()  {
	fmt.Println("Usage:")
	fmt.Println("getbalance -address [ADDRESS] -Get balance of ADDRESS")
	fmt.Println("printchain -Print all the blocks of the blockchain")
	fmt.Println("createblcokchain -address ADDRESS -Create a blockchain and send genesis block reward to Address")
	fmt.Println("send -from [FROM] -to [TO] -amount [AMOUNT] -Send AMOUNT fo coins from FROM addree to TO")
}

//验证命令参数
func (cli *CLI)validataArgs() {
	if len(os.Args)<2{
		cli.printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) printChain()  {
	var bc = block.GetBlockchain("")
	defer bc.Db.Close()

	var bci = bc.Iterator()

	for{
		var bc = bci.Next()

		fmt.Printf("Prev. hash: %x\n",bc.PrevBlockHash)
		fmt.Printf("Hash: %x\n",bc.Hash)
		var pow = block.NewProofOfWork(bc)
		fmt.Printf("Pow: %s\n",strconv.FormatBool(pow.Validate()))
		fmt.Println()

		if len(bc.PrevBlockHash) == 0{
			break
		}
	}
}

func (cli *CLI)Send(from,to string,amount int){
	var bc = block.GetBlockchain(from)
	defer bc.Db.Close()

	var tx = block.NewUTXOTransation(from,to,amount,bc)
	bc.MinBlcok([]*block.Transation{tx})
	fmt.Println("Success!")

}


//命令参数注册与解析
func (cli *CLI)Run()  {
	cli.validataArgs()

	//注册命令
	var getBalanceCmd = flag.NewFlagSet("getbalance",flag.ExitOnError)
	var printChainCmd = flag.NewFlagSet("printchian",flag.ExitOnError)
	var createBlockchainCmd = flag.NewFlagSet("createblockchain",flag.ExitOnError)
	var sendCmd = flag.NewFlagSet("send",flag.ExitOnError)

	var getBalanceAddress =
		getBalanceCmd.String("address","","The address to get balance for")
	var createBlockchainAddress =
		createBlockchainCmd.String("address","","The address to send genesis blcok reward to")
	var sendFrom = sendCmd.String("from","","Source wallet address")
	var sendTo = sendCmd.String("to","","Destination wallet address")
	var sendAmount = sendCmd.Int("amount",0,"Amount to send")

	//解析命令
	switch os.Args[1] {
	case "getbalance":
		var err = getBalanceCmd.Parse(os.Args[2:])
		utils.LogErr(err)

	case "printchain":
		var err = printChainCmd.Parse(os.Args[2:])
		utils.LogErr(err)

	case "createblockchain":
		var err = createBlockchainCmd.Parse(os.Args[2:])
		utils.LogErr(err)

	case "send":
		var err = sendCmd.Parse(os.Args[2:])
		utils.LogErr(err)

	default:
		cli.printUsage()
		os.Exit(1)
	}

	if getBalanceCmd.Parsed(){
		if *getBalanceAddress == ""{
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}


	if createBlockchainCmd.Parsed(){
		if *createBlockchainAddress == ""{
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed(){
		cli.printChain()
	}

	if sendCmd.Parsed(){
		if *sendFrom == ""||*sendTo==""||*sendAmount <=0{
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.Send(*sendFrom,*sendTo,*sendAmount)
	}
}








