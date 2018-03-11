package block

import (
	"github.com/boltdb/bolt"
	"utils"
)

//区块链迭代器
type BlockchainIterator struct {
	currentHash []byte
	db          *bolt.DB
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