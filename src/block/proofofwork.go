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

const targetBits = 2

type ProofoOfWork struct {
	block     *Block
	target    *big.Int
}

//根据提供的尚未挖掘的块的信息和难度值新建一个工作量证明对象
func NewProofOfWork(b* Block) *ProofoOfWork  {
	var target = big.NewInt(1)
	target.Lsh(target,uint(256-targetBits))  //左移获得难度值

	var pow = &ProofoOfWork{b,target}

	return pow
}

//将随机数加入到提供的块信息中构造计算对象
func (pow *ProofoOfWork)prepareData (nonce int) []byte  {

	var data = bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.HashTransaction(),
			utils.IntToHex(pow.block.TimeStamp),
			utils.IntToHex(int64(targetBits)),
			utils.IntToHex(int64(nonce)),
		},
		[]byte{},
	)
	return data
}

//执行工作量证明，也就是挖掘任务
func (pow *ProofoOfWork) Run()(int,[]byte)  {

	var hashInt big.Int
	var hash [32]byte
	var nonce = 0  //随机数从零开始

	fmt.Printf("Mining a new block")

	for nonce < maxNonce{
		var data = pow.prepareData(nonce)

		hash = sha256.Sum256(data)   //将准备的数据转化为256字节hash值

		fmt.Printf("\r%x",hash)
		hashInt.SetBytes(hash[:])  //将hash转化为大端无符号整型

		if hashInt.Cmp(pow.target) == -1{ //如果比较得出hashInt比目标难度值小就停止，否则随机数+1继续
			break
		} else {
			nonce++
		}
	}

	fmt.Printf("\n\n")
	return nonce,hash[:]   //返回计算得到的结果，hash[:]作为新的块的hash值
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





