package block

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"utils"
	"log"
	"crypto/ecdsa"
	"strings"
	"crypto/rand"
	"crypto/elliptic"
	"math/big"
)

const subsidy = 10   //奖励数额

//代表一次比特币交易
type Transaction struct {
	ID      []byte
	Vin     []TXInput
	Vout    []TXOutput
}


//检查交易是否是coinbase交易
func (tx Transaction)IsCoinBose()bool  {
	return (len(tx.Vin) == 1) && (len(tx.Vin[0].Txid) == 0) && (tx.Vin[0].Vout == -1)
}

//返回一个序列化的交易
func (tx *Transaction)Serialize() []byte  {
	var encoded bytes.Buffer

	var enc = gob.NewEncoder(&encoded)
	var err = enc.Encode(tx)
	utils.LogErr(err)

	return encoded.Bytes()

}

//返回交易的哈希
func (tx *Transaction) Hash()[]byte  {
	var hash [32]byte

	var txCopy = *tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:]
}

//对每个交易的输入进行签名
//参数为私钥和前一个交易的map
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey,prevTXs map[string]Transaction)  {
	if tx.IsCoinBose(){
		return
	}

	for _,vin :=range tx.Vin{
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil{
			log.Panic("ERROR:Previous transaction is not correct")
		}
	}

	var txCopy = tx.TrimmedCopy()

	/*
	迭代每个输入
	在每个输入中，Signature 被设置为 nil (仅仅是一个双重检验)，PubKey 被设置为所引用输出的 PubKeyHash。
	现在，除了当前交易，其他所有交易都是“空的”，
	也就是说他们的 Signature 和 PubKey 字段被设置为 nil。
	因此，输入是被分开签名的，尽管这对于我们的应用并不十分紧要，但是比特币允许交易包含引用了不同地址的输入。
	*/
	for inID,vin:=range txCopy.Vin{
		var prevTx = prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash

		/*
		Hash 方法对交易进行序列化，并使用 SHA-256 算法进行哈希。哈希后的结果就是我们要签名的数据。
		在获取完哈希，我们应该重置 PubKey 字段，以便于它不会影响后面的迭代。
		*/
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		/*
		关键点，签名
		们通过 privKey 对 txCopy.ID 进行签名。
		一个 ECDSA 签名就是一对数字，我们对这对数字连接起来，并存储在输入的 Signature 字段。
		*/
		r,s,err:=ecdsa.Sign(rand.Reader,&privKey,txCopy.ID)
		utils.LogErr(err)
		var signature = append(r.Bytes(),s.Bytes()...)

		tx.Vin[inID].Signature = signature
	}

}

//创建用于签名的修建交易副本，副本包含所有的输入和输出，但是TXInput.Signature和TXInput.PubKey被设置为nil
func (tx *Transaction)TrimmedCopy()Transaction  {
	var inputs  []TXInput
	var outputs []TXOutput

	for _,vin := range tx.Vin{
		inputs = append(inputs,TXInput{vin.Txid,vin.Vout,nil,nil})
	}

	for _,vout:=range tx.Vout{
		outputs = append(outputs,TXOutput{vout.Value,vout.PubKeyHash})
	}

	var txCopy = Transaction{tx.ID,inputs,outputs}

	return txCopy
}


//将交易转化为可读的字符串
func (tx Transaction)String()string  {
	var lines []string

	lines = append(lines,fmt.Sprintf("--- Transaction %x:",tx.ID))

	for i,input:=range tx.Vin{
		lines = append(lines,fmt.Sprintf("      Input %d:",i))
		lines = append(lines,fmt.Sprintf("      TXID %x:",input.Txid))
		lines = append(lines,fmt.Sprintf("      Out %d:",input.Vout))
		lines = append(lines,fmt.Sprintf("      Signature %x:",input.Signature))
		lines = append(lines,fmt.Sprintf("      PubKey %x:",input.PubKey))
	}

	for i,output := range tx.Vout{
		lines = append(lines,fmt.Sprintf("      Output %d",i))
		lines = append(lines,fmt.Sprintf("      Value %d",output.Value))
		lines = append(lines,fmt.Sprintf("      Output %x",output.PubKeyHash))
	}
	return strings.Join(lines,"\n")
}

//验证签名
func (tx *Transaction)Verify(prevTXs map[string]Transaction) bool  {
	if tx.IsCoinBose(){
		return true
	}

	for _,vin := range tx.Vin{
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil{
			log.Panic("ERROR: Previous transaction is correct")
		}
	}

	var txCopy = tx.TrimmedCopy()
	var curve = elliptic.P256()


	//检查每个输入中的签名
	for inID,vin := range tx.Vin{
		var prevTx = prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTx.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inID].PubKey = nil

		var r = big.Int{}
		var s = big.Int{}

		var sigLen = len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen/2)])
		s.SetBytes(vin.Signature[(sigLen/2):])

		var x = big.Int{}
		var y = big.Int{}

		var keyLen = len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen/2)])
		y.SetBytes(vin.PubKey[(keyLen/2):])

		var rawPubKey = ecdsa.PublicKey{curve,&x,&y}
		if ecdsa.Verify(&rawPubKey,txCopy.ID,&r,&s)==false{
			return false
		}
	}
	return true
}

//新建一个coinbase交易
func NewCoinbaseTX(to,data string) *Transaction  {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'",to)
	}

	var txin = TXInput{[]byte{},-1,nil,[]byte(data)}
	var txout = NewTXOutput(subsidy,to)
	var tx = Transaction{nil,[]TXInput{txin},[]TXOutput{*txout}}
	tx.ID = tx.Hash()

	return &tx
}

//新建一个交易
func NewUTXOTransaction(from,to string,amount int,bc *Blockchain) *Transaction {
	var inputs  []TXInput
	var outputs  []TXOutput

	var wallets,err = GetWallets()
	utils.LogErr(err)
	var wallet = wallets.GetWallet(from)
	var pubKeyHash = HashPubKey(wallet.PublicKey)


	var acc,validOutputs = bc.FindSpendableOutput(pubKeyHash,amount)

	if acc<amount{
		log.Panic("Error: Not enough funds")
	}

	for txid,outs:=range validOutputs{
		var txID,err = hex.DecodeString(txid)
		utils.LogErr(err)
		for _,out:=range outs{
			var input = TXInput{txID,out,nil,wallet.PublicKey}
			inputs = append(inputs,input)
		}
	}

	//build a list of outputs
	outputs = append(outputs,*NewTXOutput(amount,to))
	if acc > amount{
		outputs = append(outputs,*NewTXOutput(acc - amount,from)) //a change
	}

	var tx = Transaction{nil,inputs,outputs}
	tx.ID = tx.Hash()
	//bc.SignTransaction(&tx,wallet.PrivateKey)

	return &tx
}










































