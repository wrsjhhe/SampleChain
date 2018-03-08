package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"utils"
	"log"
)

const subsidy = 10   //奖励数额

//交易输出
type TXOutput struct{
	Value          int      //储存了“币”数量
	ScriptPubKey   string   //用户定义的钱包地址
}

//交易输入
type TXInput struct {
	Txid        []byte     //交易ID
	Vout        int        //该输出在这笔交易中所有输出的索引（一笔交易可能有多个输出）
	ScriptSig   string	   //脚本，提供了可作用于一个输出的ScriptPubkey的数据
}


type Transation struct {
	ID      []byte
	Vin     []TXInput
	Vout    []TXOutput
}


//检查交易是否是coinbase交易
func (tx Transation)IsCoinBose()bool  {
	return (len(tx.Vin) == 1) && (len(tx.Vin[0].Txid) == 0) && (tx.Vin[0].Vout == -1)
}

//设置交易ID
func (tx *Transation)SetID()  {
	var enconded bytes.Buffer
	var hash     [32]byte

	var enc = gob.NewEncoder(&enconded)
	var err = enc.Encode(tx)
	utils.LogErr(err)
	hash = sha256.Sum256(enconded.Bytes())
	tx.ID = hash[:]
}

//检查解锁数据是否能启动交易
func (in *TXInput)CanUnlockOutputWith(unlockingData string)bool  {
	return in.ScriptSig == unlockingData
}

//检查输出是否能被解锁数据解锁
func (out *TXOutput)CanBeUnlockedWith(unlockingData string)bool  {
	return out.ScriptPubKey == unlockingData
}

//新建一个coinbase交易
func NewCoinbaseTX(to,data string) *Transation  {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'",to)
	}

	var txin = TXInput{[]byte{},-1,data}
	var txout = TXOutput{subsidy,to}
	var tx = Transation{nil,[]TXInput{txin},[]TXOutput{txout}}
	tx.SetID()

	return &tx
}

//新建一个交易
func NewUTXOTransation(from,to string,amount int,bc *Blockchain) *Transation {
	var inputs  []TXInput
	var outputs  []TXOutput

	var acc,validOutputs = bc.FindSpendabbleOutput(from,amount)

	if acc<amount{
		log.Panic("Error: Not enough funds")
	}

	for txid,outs:=range validOutputs{
		var txID,err = hex.DecodeString(txid)
		utils.LogErr(err)
		for _,out:=range outs{
			var input = TXInput{txID,out,from}
			inputs = append(inputs,input)
		}
	}

	outputs = append(outputs,TXOutput{amount,to})
	if acc > amount{
		outputs = append(outputs,TXOutput{acc - amount,from})
	}

	var tx = Transation{nil,inputs,outputs}
	tx.SetID()

	return &tx
}










































