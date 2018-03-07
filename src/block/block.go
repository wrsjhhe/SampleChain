package block

import (
	"time"
	"bytes"
	"encoding/gob"
	"log"
)

type Block struct {
	TimeStamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce		  int
}

//序列化区块,因为在boltdb中值只能是[]byte，而我们想要存的是结构
func (b* Block)Serialize() []byte  {
	var result bytes.Buffer //储存序列化之后的数据

	var encoder = gob.NewEncoder(&result) //序列化器

	var err = encoder.Encode(b)
	if err!=nil{
		log.Panic(err)
	}

	return  result.Bytes()
}

//反序列化
func DeserializeBlock(d []byte) *Block {
	var block Block

	var decoder = gob.NewDecoder(bytes.NewReader(d))
	var err = decoder.Decode(&block)

	if err!=nil{
		log.Panic(err)
	}
	return &block
}

func NewBlock(data string,prevBlockHash []byte) *Block  {

	var block = &Block{
		time.Now().Unix(),
		[]byte(data),
		prevBlockHash,
		[]byte{},
		0}

	var pow = NewProofOfWork(block)

	var nonce,hash = pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func NewGenesisBlock() *Block {
	return NewBlock("Genesis Block",[]byte{})

}






