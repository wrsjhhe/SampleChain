package block

import (
	"github.com/boltdb/bolt"
	"fmt"
	"os"
	"encoding/hex"
	"utils"
	"crypto/ecdsa"
	"bytes"
	"errors"
	"log"
)

const dbFile  = "blockchain_%s.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout of banks"

//区块链，保存一系列的区块
type Blockchain struct {
	Tip []byte   //数据库中储存的最后一个块的哈希
	Db  *bolt.DB
}


//创建新的链
func CreateBlockchain(address,nodeID string)*Blockchain  {

	var dbFile = fmt.Sprintf(dbFile,nodeID)

	if dbExists(dbFile){
		fmt.Println("Blockchain already exists")
		os.Exit(1)
	}

	var tip []byte

	var cbtx = NewCoinbaseTX(address,genesisCoinbaseData)
	var genesis = NewGenesisBlock(cbtx)

	//打开一个db文件标准做法，文件不存在不会返回错误
	var db ,err = bolt.Open(dbFile,0600,nil)
	utils.LogErr(err)

	err = db.Update(func(tx *bolt.Tx) error { //读写事物

		var b,err = tx.CreateBucket([]byte(blocksBucket))
		utils.LogErr(err)

		err = b.Put(genesis.Hash,genesis.Serialize())
		utils.LogErr(err)

		err = b.Put([]byte("l"),genesis.Hash)
		utils.LogErr(err)

		tip = genesis.Hash

		return nil
	})
	utils.LogErr(err)

	var bc = Blockchain{tip,db}
	return &bc

}


//返回现有的区块链
func GetBlockchain(nodeID string)*Blockchain {
	var dbFile = fmt.Sprintf(dbFile,nodeID)

	if dbExists(dbFile) == false{
		fmt.Println("No existing blockchain found.Create one first")
		os.Exit(1)
	}

	var tip []byte
	//打开一个db文件标准做法，文件不存在不会返回错误
	var db ,err = bolt.Open(dbFile,0600,nil)
	utils.LogErr(err)

	err = db.Update(func(tx *bolt.Tx) error { //读写事物
		var b= tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	utils.LogErr(err)

	var bc = Blockchain{tip,db}
	return &bc
}

//保存块到区块链
func (bc *Blockchain)AddBlock(block *Block)  {
	var err = bc.Db.Update(func(tx *bolt.Tx) error {
		var b = tx.Bucket([]byte(blocksBucket))
		var blockInDb = b.Get(block.Hash)

		if blockInDb == nil{
			return nil
		}

		var blockData = block.Serialize()
		var err = b.Put(block.Hash,blockData)
		utils.LogErr(err)

		var lastHash = b.Get([]byte("l"))
		var lastBlockData = b.Get(lastHash)
		var lastBlock = DeserializeBlock(lastBlockData)

		if block.Height > lastBlock.Height{
			err = b.Put([]byte("l"),block.Hash)
			utils.LogErr(err)
			bc.Tip = block.Hash
		}
		return nil
	})
	utils.LogErr(err)
}

func (bc *Blockchain)GetBlockHashes()[][]byte  {
	var blocks [][]byte
	var bci = bc.Iterator()

	for{
		var block = bci.Next()

		blocks = append(blocks,block.Hash)
		if len(block.PrevBlockHash) == 0{
			break
		}
	}
	return blocks
}

//通过提供的交易数据来挖掘新块
func (bc *Blockchain)MineBlock(transations []*Transaction) *Block {
	var lastHash   []byte
	var lastHeight int

	for _,tx:=range transations{
		if bc.VerifyTransaction(tx)!=true{
			log.Panic("ERROR,Invalid transaction")
		}
	}


	var err = bc.Db.View(func(tx *bolt.Tx) error {
		var b = tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))

		var blockData = b.Get(lastHash)
		var block = DeserializeBlock(blockData)

		lastHeight = block.Height
		return nil
	})

	utils.LogErr(err)

	var newBlock = NewBlock(transations,lastHash,lastHeight+1)

	err = bc.Db.Update(func(tx *bolt.Tx) error {
		var b = tx.Bucket([]byte(blocksBucket))
		var err = b.Put(newBlock.Hash,newBlock.Serialize())

		utils.LogErr(err)

		err = b.Put([]byte("l"),newBlock.Hash)
		utils.LogErr(err)

		bc.Tip = newBlock.Hash
		return nil
	})
	return newBlock
}

//通过交易ID返回一个交易
func (bc *Blockchain)FindTransaction(ID []byte)(Transaction,error)  {
	var bci = bc.Iterator()

	for{
		var block = bci.Next()

		for _,tx:=range block.Transactions{
			if bytes.Compare(tx.ID,ID) == 0{
				return *tx,nil
			}
		}

		if len(block.PrevBlockHash)==0{
			break
		}
	}
	return Transaction{},errors.New("Transaction is not found")

}

//找到包含未花费输出的交易
func (bc *Blockchain)FindUnspentTransactions(pubKeyHash []byte) []Transaction  {
	var unspentTXs []Transaction
	var spentTXOs = make(map[string][]int)
	var bci = bc.Iterator()

	for{
		var block = bci.Next()

		for _,tx:=range block.Transactions{
			var txID = hex.EncodeToString(tx.ID)

			Outputs:
				for outIdx,out :=range tx.Vout{
					//输出花费了没有?
					if spentTXOs[txID] != nil{
						for _,spendOut:=range spentTXOs[txID]{
							if spendOut == outIdx{
								continue Outputs
							}
						}
					}

					if out.IsLockedWithKey(pubKeyHash){
						unspentTXs = append(unspentTXs,*tx)
					}
				}

				if tx.IsCoinBose() == false{
					for _,in := range tx.Vin{
						if in.UsesKey(pubKeyHash){
							var inTxID = hex.EncodeToString(in.Txid)
							spentTXOs[inTxID] = append(spentTXOs[inTxID],in.Vout)
						}
					}
				}
		}

		if len(block.PrevBlockHash) == 0{
			break
		}
	}
	return  unspentTXs
}

//找到并返回所有非花费输出,返回的是移除了花费了的输出
func (bc *Blockchain)FindUTXO() map[string]TXOutputs  {
	var UTXO = make(map[string]TXOutputs)

	var spentTXOs = make(map[string][]int)
	var bci = bc.Iterator()

	for{
		var blck = bci.Next()

		for _,tx:=range blck.Transactions{
			var txID = hex.EncodeToString(tx.ID)

			Outputs:
				for outIdx,out:=range tx.Vout{
					//是否输出被花费了
					if spentTXOs[txID] !=nil{
						for _,spentOutIdx:=range spentTXOs[txID]{
							if spentOutIdx == outIdx{
								continue Outputs
							}
						}
					}

					var outs = UTXO[txID]
					outs.Outputs = append(outs.Outputs,out)
					UTXO[txID] = outs
				}

				if tx.IsCoinBose() == false{
					for _,in := range tx.Vin{
						var inTxId = hex.EncodeToString(in.Txid)
						spentTXOs[inTxId] = append(spentTXOs[inTxId],in.Vout)
					}
				}
		}

		if len(blck.PrevBlockHash) == 0{
			break
		}
	}
	return UTXO
}

func (bc *Blockchain)FindSpendableOutput(pubKeyHash []byte,amount int)(int,map[string][]int)  {
	var unspentOutputs = make(map[string][]int)
	var unspentTXs = bc.FindUnspentTransactions(pubKeyHash)
	var accumulated = 0

	Work:
		for _,tx:=range unspentTXs{
			var txID = hex.EncodeToString(tx.ID)

			for outIdx,out :=range tx.Vout{
				if out.IsLockedWithKey(pubKeyHash)&& accumulated < amount{
					accumulated +=out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID],outIdx)

					if accumulated >=amount{
						break Work
					}
				}
			}
		}
		return accumulated , unspentOutputs
}

//对一个交易的输入签名
func (bc *Blockchain)SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey)  {
	var prevTXs = make(map[string]Transaction)

	for _,vin:= range tx.Vin{
		prevTX,err:=bc.FindTransaction(vin.Txid)
		utils.LogErr(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	tx.Sign(privKey,prevTXs)
}

func (bc *Blockchain)VerifyTransaction(tx *Transaction)bool  {
	var prevTXs = make(map[string]Transaction)

	for _,vin := range tx.Vin{
		var prevTX,err = bc.FindTransaction(vin.Txid)
		utils.LogErr(err)

		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	return tx.Verify(prevTXs)
}


//从最后一个块向前遍历链
func (bc *Blockchain) Iterator() *BlockchainIterator  {
	var bci = &BlockchainIterator{bc.Tip,bc.Db}
	return bci
}

//返回最后一个块的高度
func (bc* Blockchain)GetBestHeight()int  {
	var lastBlock Block

	var err = bc.Db.View(func(tx *bolt.Tx) error {
		var b = tx.Bucket([]byte(blocksBucket))
		var lastHash = b.Get([]byte("l"))
		var blockData = b.Get(lastHash)
		lastBlock = *DeserializeBlock(blockData)

		return nil
	})
	utils.LogErr(err)

	return lastBlock.Height
}

func (bc *Blockchain)GetBlock(blockHash []byte)(Block,error)  {
	var block Block

	var err = bc.Db.View(func(tx *bolt.Tx) error {
		var b = tx.Bucket([]byte(blocksBucket))

		var blockData = b.Get(blockHash)
		if blockData == nil{
			return errors.New("Block is not found.")
		}

		block = *DeserializeBlock(blockData)
		return nil
	})
	utils.LogErr(err)
	return block ,nil
}

func dbExists(dbFile string)bool  {
	if _,err :=os.Stat(dbFile);os.IsNotExist(err){
		return false
	}
	return true
}
