package block

import (
	"time"
	"bytes"
	"encoding/gob"
	"crypto/sha256"
	"utils"
)

type Block struct {
	TimeStamp     int64
	Transactions  []*Transaction
	PrevBlockHash []byte
	Hash          []byte  //比特币用的Merkle树
	Nonce		  int
}

//序列化区块,因为在boltdb中值只能是[]byte，而我们想要存的是结构
func (b* Block)Serialize() []byte  {
	var result bytes.Buffer //储存序列化之后的数据

	var encoder = gob.NewEncoder(&result) //序列化器

	var err = encoder.Encode(b)
	utils.LogErr(err)

	return  result.Bytes()
}

//反序列化
func DeserializeBlock(d []byte) *Block {
	var block Block

	var decoder = gob.NewDecoder(bytes.NewReader(d))
	var err = decoder.Decode(&block)

	utils.LogErr(err)
	return &block
}

//将块中的所有交易转化为hash
func (b *Block)HashTransaction()[]byte  {
	var txHashes    [][]byte
	var txHash      [32]byte
	for _, tx := range b.Transactions{
		txHashes = append(txHashes,tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes,[]byte{}))

	return txHash[:]
}


func NewBlock(transation []*Transaction,prevBlockHash []byte) *Block  {

	var block = &Block{
		time.Now().Unix(),
		transation,
		prevBlockHash,
		[]byte{},
		0}

	var pow = NewProofOfWork(block)

	var nonce,hash = pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func NewGenesisBlock(coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{coinbase},[]byte{})

}






