package block

import (
	"fmt"
	"bytes"
	"encoding/gob"
	"utils"
	"net"
	"io"
	"encoding/hex"
	"io/ioutil"
)

const protocol = "tcp"
const nodeVersion = 1
const commandLength = 12   //

var nodeAddress string
var miningAddress string
var KnownNodes = []string{"localhost:3000"}
var blocksInTransit = [][]byte{}
var mempool = make(map[string]Transaction)

type addr struct{
	AddrList    []string
}

type block struct{
	AddrFrom    string
	Block       []byte
}

type getblocks struct{
	AddrFrom    string
}

type getdata struct{
	AddrFrom   string
	Type       string
	ID         []byte
}

type inv struct{
	AddrFrom  string
	Type      string
	Items     [][]byte
}

type tx struct {
	AddrFrom      string
	Transaction   []byte
}

type verzion struct {
	Version     int
	BestHeight  int
	AddrFrom    string
}

func commandToBytes(command string)[]byte  {
	var bytes  [commandLength]byte

	for i,c:=range command{
		bytes[i] = byte(c)
	}
	return bytes[:]
}



func bytesToCommand(bytes  []byte)string  {
	var command []byte

	for _,b := range bytes{
		if b!=0x0{
			command = append(command,b)
		}
	}
	return fmt.Sprintf("%s",command)
}

func extractCommand(request []byte)[]byte  {
	return request[:]
}

func requestBlocks()  {
	for _,node :=range KnownNodes{
		sendGetBlocks(node)
	}
}

func sendAddr(address string)  {
	var nodes = addr{KnownNodes}
	nodes.AddrList = append(nodes.AddrList,nodeAddress)
	var payload = gobEncode(nodes)
	var request = append(commandToBytes("addr"),payload...)

	sendData(address,request)
}

func sendBlock(addr string,b *Block)  {
	var data = block{nodeAddress,b.Serialize()}
	var payload = gobEncode(data)
	var request = append(commandToBytes("block"),payload...)

	sendData(addr,request)
}

func sendData(addr string,data []byte)  {
	var conn,err = net.Dial(protocol,addr)
	if err!=nil{
		fmt.Printf("%s is not available\n",addr)
		var updateNodes  []string

		for _,node := range KnownNodes{
			if node!=addr{
				updateNodes = append(updateNodes,node)
			}
		}
		KnownNodes = updateNodes
		return
	}
	defer conn.Close()

	_,err = io.Copy(conn,bytes.NewReader(data))
	utils.LogErr(err)
}

func sendInv(address ,kind string,items [][]byte)  {
	var inventory = inv{nodeAddress,kind,items}
	var payload = gobEncode(inventory)
	var request = append(commandToBytes("inv"),payload...)

	sendData(address,request)
}

func sendGetBlocks(address string)  {
	var payload = gobEncode(getblocks{nodeAddress})
	var request = append(commandToBytes("getblocks"),payload...)

	sendData(address,request)
}

func sendGetData(address ,kind string,id []byte)  {
	var payload = gobEncode(getblocks{nodeAddress})
	var request = append(commandToBytes("getblocks"),payload...)

	sendData(address,request)
}

func SendTx(addr string,tnx *Transaction)  {
	var data = tx{nodeAddress,tnx.Serialize()}
	var payload = gobEncode(data)
	var request = append(commandToBytes("tx"),payload...)

	sendData(addr,request)
}

func sendVersion(addr string,bc *Blockchain)  {
	var bestHeight = bc.GetBestHeight()
	var payload = gobEncode(verzion{nodeVersion,bestHeight,nodeAddress})

	var request = append(commandToBytes("version"),payload...)

	sendData(addr,request)
}

func handleAddr(request []byte)  {
	var buff bytes.Buffer
	var payload  addr

	buff.Write(request[commandLength:])
	var dec = gob.NewDecoder(&buff)
	var err = dec.Decode(&payload)
	utils.LogErr(err)

	KnownNodes = append(KnownNodes,payload.AddrList...)
	fmt.Printf("There are %d known nodes now!\n",len(KnownNodes))
	requestBlocks()
}

func handleBlock(request []byte,bc *Blockchain)  {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	var dec = gob.NewDecoder(&buff)
	var err = dec.Decode(&payload)
	utils.LogErr(err)

	var blockData = payload.Block
	var block = DeserializeBlock(blockData)

	fmt.Println("Received a new block!")
	bc.AddBlock(block)

	fmt.Printf("Added block %x\n",block.Hash)

	if len(blocksInTransit) > 0{
		var blockHash = blocksInTransit[0]
		sendGetData(payload.AddrFrom,"block",blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else{
		var UTXOSet = UTXOSet{bc}
		UTXOSet.Reindex()
	}
}

func handleInv(request []byte,bc *Blockchain)  {
	var buff bytes.Buffer
	var payload  inv

	buff.Write(request[commandLength:])
	var dec = gob.NewDecoder(&buff)
	var err = dec.Decode(&payload)
	utils.LogErr(err)

	fmt.Printf("Reveived inventory with %d %s\n",len(payload.Items),payload.Type)

	if payload.Type == "block"{
		blocksInTransit = payload.Items

		var blockHash = payload.Items[0]
		sendGetData(payload.AddrFrom,"block",blockHash)

		var newInTransit = [][]byte{}
		for _,b :=range blocksInTransit{
			if bytes.Compare(b,blockHash)!=0{
				newInTransit = append(newInTransit,b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx"{
		var txID = payload.Items[0]

		if mempool[hex.EncodeToString(txID)].ID == nil{
			sendGetData(payload.AddrFrom,"tx",txID)
		}
	}
}

func handleGetBlocks(request []byte,bc *Blockchain)  {
	var buff bytes.Buffer
	var payload getblocks

	buff.Write(request[commandLength:])
	var dec = gob.NewDecoder(&buff)
	var err = dec.Decode(&payload)
	utils.LogErr(err)

	var blocks = bc.GetBlockHashes()
	sendInv(payload.AddrFrom,"block",blocks)
}

func handleGetData(request []byte,bc *Blockchain)  {
	var buff bytes.Buffer
	var payload getdata

	buff.Write(request[commandLength:])
	var dec = gob.NewDecoder(&buff)
	var err = dec.Decode(&payload)
	utils.LogErr(err)

	if payload.Type == "block"{
		var block,err = bc.GetBlock([]byte(payload.ID))
		utils.LogErr(err)

		sendBlock(payload.AddrFrom,&block)
	}

	if payload.Type == "tx"{
		var txID = hex.EncodeToString(payload.ID)
		var tx = mempool[txID]

		SendTx(payload.AddrFrom,&tx)
	}
}

func handleTx(request []byte,bc *Blockchain)  {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:])
	var dec = gob.NewDecoder(&buff)
	var err = dec.Decode(&payload)
	utils.LogErr(err)

	var txData = payload.Transaction
	var tx = DeserializeTransaction(txData)
	mempool[hex.EncodeToString(tx.ID)] = tx

	if nodeAddress == KnownNodes[0]{
		for _,node :=range KnownNodes{
			sendInv(node,"tx",[][]byte{tx.ID})
		}
	}else{
		if len(mempool) >=2 && len(miningAddress) >0{
			MineTransactions:
				var txs []*Transaction

				for id:=range mempool{
					var tx = mempool[id]
					if bc.VerifyTransaction(&tx){
						txs = append(txs,&tx)
					}
				}

				if len(txs) == 0{
					fmt.Println("All transactions are invalid! Waiting for new ones...")
					return
				}

				var cbTx = NewCoinbaseTX(miningAddress,"")
				txs = append(txs,cbTx)

				var newBlock = bc.MineBlock(txs)
				var UTXOSet = UTXOSet{bc}
				UTXOSet.Reindex()

				fmt.Println("New block is mined")

				for _,tx:=range txs{
					var txID = hex.EncodeToString(tx.ID)
					delete(mempool,txID)
				}

				for _,node:=range KnownNodes{
					if node != nodeAddress{
						sendInv(node,"block",[][]byte{newBlock.Hash})
					}
				}

				if len(mempool)>0{
					goto MineTransactions
				}
		}
	}
}

func handleVersion(request []byte,bc *Blockchain)  {
	var buff bytes.Buffer
	var payload verzion

	buff.Write(request[commandLength:])
	var dec = gob.NewDecoder(&buff)
	var err = dec.Decode(&payload)
	utils.LogErr(err)

	var myBestHeight = bc.GetBestHeight()
	var foreignerBestHeight = payload.BestHeight

	if myBestHeight < foreignerBestHeight{
		sendGetBlocks(payload.AddrFrom)
	}else if myBestHeight > foreignerBestHeight{
		sendVersion(payload.AddrFrom,bc)
	}

	if !nodeIsKnown(payload.AddrFrom){
		KnownNodes = append(KnownNodes,payload.AddrFrom)
	}
}

func handleConnection(conn net.Conn,bc *Blockchain)  {
	var request,err = ioutil.ReadAll(conn)
	utils.LogErr(err)

	var command = bytesToCommand(request[:commandLength])
	fmt.Printf("Received %s command\n",command)

	switch command {
	case "addr":
		handleAddr(request)
	case "block":
		handleBlock(request,bc)
	case "inv":
		handleInv(request,bc)
	case "getblocks":
		handleGetBlocks(request,bc)
	case "getdata":
		handleGetData(request,bc)
	case "tx":
		handleTx(request,bc)
	case "version":
		handleVersion(request,bc)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
}

func StartServer(nodeID,minerAddress string)  {
	var nodeAddress = fmt.Sprintf("localhost:%s",nodeID)
	miningAddress = minerAddress

	ln,err := net.Listen(protocol,nodeAddress)
	utils.LogErr(err)

	defer ln.Close()

	var bc = GetBlockchain(nodeID)

	if nodeAddress != KnownNodes[0]{
		sendVersion(KnownNodes[0],bc)
	}

	for{
		conn,err:=ln.Accept()
		utils.LogErr(err)

		go handleConnection(conn,bc)
	}
}

func gobEncode(data interface{})[]byte  {
	var buff bytes.Buffer

	var enc = gob.NewEncoder(&buff)
	var err = enc.Encode(data)

	utils.LogErr(err)

	return buff.Bytes()
}

func nodeIsKnown(addr string)bool  {
	for _,node :=range KnownNodes{
		if node == addr{
			return true
		}
	}
	return false
}






























