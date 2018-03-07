package block

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"utils"
)


var (
	maxNonce = math.MaxInt64
)

const targetBits = 24

type ProofoOfWork struct {
	block     *Block
	target    *big.Int
}


func NewProofOfWork(b* Block) *ProofoOfWork  {
	var target = big.NewInt(1)
	target.Lsh(target,uint(256-targetBits))

	var pow = &ProofoOfWork{b,target}

	return pow
}

func (pow *ProofoOfWork)prepareData (nonce int) []byte  {

	var data = bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			utils.IntToHex(pow.block.TimeStamp),
			utils.IntToHex(int64(targetBits)),
			utils.IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

//执行pow
func (pow *ProofoOfWork) Run()(int,[]byte)  {

	var hashInt big.Int
	var hash [32]byte
	var nonce = 0

	fmt.Printf("Mining the block containing \"%s\"\n",pow.block.Data)

	for nonce < maxNonce{
		var data = pow.prepareData(nonce)

		hash = sha256.Sum256(data)

		fmt.Printf("\r%x",hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1{
			break
		} else {
			nonce++
		}
	}

	fmt.Printf("\n\n")
	return nonce,hash[:]
}

//验证块的POW
func (pow *ProofoOfWork) Validate()bool  {
	var hashInt big.Int

	var data = pow.prepareData(pow.block.Nonce)

	var hash = sha256.Sum256(data)

	hashInt.SetBytes(hash[:])

	var isValid = hashInt.Cmp(pow.target) == -1

	return isValid
}





