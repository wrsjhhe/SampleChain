package block

import (
	"bytes"
	"utils"
	"encoding/gob"
)

//交易输出
type TXOutput struct{
	Value          int      //储存了“币”数量
	PubKeyHash     []byte   //公钥哈希
}

type TXOutputs struct {
	Outputs    []TXOutput
}


//检查是否提供的公钥哈希被用于锁定输出
func (out *TXOutput)IsLockedWithKey(pubKeyHash []byte)bool  {
	return bytes.Compare(out.PubKeyHash,pubKeyHash) == 0
}

//锁定一个输出，当我们给某人发送币时，我们只知道他的地址，
//因此这个函数使用一个地址作为唯一参数，
//然后地址会被解码，从中提取公钥哈希保存在PubKeyHash字段
func (out *TXOutput)Lock(address []byte)  {
	var pubKeyHash = utils.Base58Decode(address)
	pubKeyHash = pubKeyHash[1:len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

//创建一个输出
func NewTXOutput(value int,address string) *TXOutput  {
	var txo = &TXOutput{value,nil}
	txo.Lock([]byte(address))
	return txo
}

//序列化TXOutputs对象到字节数组
func (outs TXOutputs)Serialize()[]byte  {
	var buff bytes.Buffer

	var enc = gob.NewEncoder(&buff)
	var err = enc.Encode(outs)
	utils.LogErr(err)

	return buff.Bytes()
}

//将字节数组反序列化为TXOutput对象
func DeserializeOutputs(data []byte)TXOutputs  {
	var outputs TXOutputs

	var dec = gob.NewDecoder(bytes.NewReader(data))
	var err = dec.Decode(&outputs)
	utils.LogErr(err)

	return outputs
}