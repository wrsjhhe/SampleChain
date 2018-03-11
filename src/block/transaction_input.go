package block

import "bytes"

//交易输入
type TXInput struct {
	Txid        []byte     //交易ID
	Vout        int        //该输出在这笔交易中所有输出的索引（一笔交易可能有多个输出）
	Signature	[]byte     //签名
	PubKey      []byte     //公钥(没有被哈希的公钥)
}

//使用指定的秘钥来解锁一个输出
func (in *TXInput)UsesKey(pubKeyHash []byte)bool  {
	var lockingHash = HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash,pubKeyHash) == 0
}