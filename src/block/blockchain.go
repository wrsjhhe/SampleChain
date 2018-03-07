package block

import (
	"log"
	"github.com/boltdb/bolt"
	"fmt"
)

const dbFIle  = "blockchain.db"
const blocksBucket = "blocks"


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


//将提供的数据转化为区块然后添加到区块链中
func (bc *Blockchain) AddBlock (data string){
	var lastHash []byte

	var err = bc.Db.View(func(tx *bolt.Tx) error {   //只读事物
		var b = tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l")) //key"l"指向的是链中最后一个块的哈希

		return nil
	})

	if err!=nil{
		log.Panic(err)
	}

	var newBlock = NewBlock(data,lastHash)

	err = bc.Db.Update(func(tx *bolt.Tx) error {  //读写事物
		var b = tx.Bucket([]byte(blocksBucket))
		var err = b.Put(newBlock.Hash,newBlock.Serialize())

		if err!=nil{
			log.Panic(err)
		}

		err = b.Put([]byte("l"),newBlock.Hash)
		if err!=nil{
			log.Panic(err)
		}

		return nil
	})

}
//遍历
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

	if err!=nil{
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash

	return block
}

//如果之前没有区块链，就从创世区块开始创建，如果有就返回现有的区块链
func NewBlockchain()*Blockchain {
	var tip []byte
	//打开一个db文件标准做法，文件不存在不会返回错误
	var db ,err = bolt.Open(dbFIle,0600,nil)
	if err!=nil{
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error { //读写事物
		var b = tx.Bucket([]byte(blocksBucket))

		if b == nil{ //说明之前还没有区块链
			fmt.Println("No existing blockchain found. Creating a new one...")
			var genesis = NewGenesisBlock()

			var b,err = tx.CreateBucket([]byte(blocksBucket))
			if err!=nil{
				log.Panic(err)
			}

			err = b.Put(genesis.Hash,genesis.Serialize())
			if err!=nil{
				log.Panic(err)
			}
			err = b.Put([]byte("l"),genesis.Hash)
			if err!=nil{
				log.Panic(err)
			}
			tip = genesis.Hash
		} else {  //说明之前存在了一个区块链

			tip = b.Get([]byte("l"))
		}
		return nil
	})
	if err!=nil{
		log.Panic(err)
	}
	var bc = Blockchain{tip,db}
	return &bc
}
