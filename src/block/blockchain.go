package block

import (
	"github.com/boltdb/bolt"
	"fmt"
	"os"
	"encoding/hex"
	"utils"
)

const dbFIle  = "blockchain.db"
const blocksBucket = "blocks"
const genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout of banks"

//区块链，保存一系列的区块
type Blockchain struct {
	Tip []byte   //数据库中储存的最后一个块的哈希
	Db  *bolt.DB
}

//区块链迭代器
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
}

//通过提供的交易数据来挖掘新块
func (bc *Blockchain)MinBlcok(transations []*Transation)  {
	var lastHash   []byte

	var err = bc.Db.View(func(tx *bolt.Tx) error {
		var b = tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})

	utils.LogErr(err)

	var newBlock = NewBlock(transations,lastHash)

	err = bc.Db.Update(func(tx *bolt.Tx) error {
		var b = tx.Bucket([]byte(blocksBucket))
		var err = b.Put(newBlock.Hash,newBlock.Serialize())

		utils.LogErr(err)

		err = b.Put([]byte("l"),newBlock.Hash)
		utils.LogErr(err)

		bc.Tip = newBlock.Hash
		return nil
	})
}

//找到包含未花费输出的交易
func (bc *Blockchain)FindUnspentTransactions(address string) []Transation  {
	var unspentTXs []Transation
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

					if out.CanBeUnlockedWith(address){
						unspentTXs = append(unspentTXs,*tx)
					}
				}

				if tx.IsCoinBose() == false{
					for _,in := range tx.Vin{
						if in.CanUnlockOutputWith(address){
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

//找到并返回所有非花费输出
func (bc *Blockchain)FindUTXO(address string) []TXOutput  {
	var UTXOs []TXOutput
	var unspentTransactions = bc.FindUnspentTransactions(address)

	for _,tx:=range unspentTransactions{
		for _,out:=range tx.Vout{
			if out.CanBeUnlockedWith(address){
				UTXOs = append(UTXOs,out)
			}
		}
	}
	return UTXOs
}

func (bc *Blockchain)FindSpendabbleOutput(address string,amount int)(int,map[string][]int)  {
	var unspentOutputs = make(map[string][]int)
	var unspentTXs = bc.FindUnspentTransactions(address)
	var accumulated = 0

	Work:
		for _,tx:=range unspentTXs{
			var txID = hex.EncodeToString(tx.ID)

			for outIdx,out :=range tx.Vout{
				if out.CanBeUnlockedWith(address)&& accumulated < amount{
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

//从最后一个块向前遍历链
func (bc *Blockchain) Iterator() *BlockchainIterator  {
	var bci = &BlockchainIterator{bc.Tip,bc.Db}
	return bci
}

//返回从tip开始的前一个区块
func (i *BlockchainIterator) Next() *Block {
	var block *Block

	var err = i.db.View(func(tx *bolt.Tx) error {
		var b = tx.Bucket([]byte(blocksBucket))
		var encodedBlock = b.Get(i.currentHash)
		block = DeserializeBlock(encodedBlock)
		return nil
	})

	utils.LogErr(err)

	i.currentHash = block.PrevBlockHash

	return block
}

func dbExists()bool  {
	if _,err :=os.Stat(dbFIle);os.IsNotExist(err){
		return false
	}
	return true
}

//返回现有的区块链
func GetBlockchain(address string)*Blockchain {
	if dbExists() == false{
		fmt.Println("No existing blockchain found.Create one first")
		os.Exit(1)
	}

	var tip []byte
	//打开一个db文件标准做法，文件不存在不会返回错误
	var db ,err = bolt.Open(dbFIle,0600,nil)
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

func CreateBlockchain(address string)*Blockchain  {
	if dbExists(){
		fmt.Println("Blockchain already exists")
		os.Exit(1)
	}

	var tip []byte
	//打开一个db文件标准做法，文件不存在不会返回错误
	var db ,err = bolt.Open(dbFIle,0600,nil)
	utils.LogErr(err)

	err = db.Update(func(tx *bolt.Tx) error { //读写事物

		var cbtx = NewCoinbaseTX(address,genesisCoinbaseData)
		var genesis = NewGenesisBlock(cbtx)

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