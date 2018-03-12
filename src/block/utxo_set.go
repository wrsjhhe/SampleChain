package block

import (
	"github.com/boltdb/bolt"
	"encoding/hex"
	"utils"
)

const utxoBucket = "chainstate"


//未花费输出集合
type UTXOSet struct {
	Blockchain  *Blockchain
}

//找到可花费的未花费输出
func (u UTXOSet)FindSpendableOutputs(pubKeyHash []byte,amount int)(int,map[string][]int)  {
	var unspentOutput = make(map[string][]int)
	var accumulated = 0
	var db = u.Blockchain.Db

	var err = db.View(func(tx *bolt.Tx) error {
		var b = tx.Bucket([]byte(utxoBucket))
		var c = b.Cursor()

		for k,v:=c.First();k!=nil;k,v = c.Next(){
			var txID = hex.EncodeToString(k)
			var outs = DeserializeOutputs(v)

			for outIDx,out :=range outs.Outputs{
				if out.IsLockedWithKey(pubKeyHash)&&accumulated < amount{
					accumulated += out.Value
					unspentOutput[txID] = append(unspentOutput[txID],outIDx)
				}
			}
		}
		return  nil
	})
	utils.LogErr(err)

	return accumulated , unspentOutput
}

func (u UTXOSet)FindUTXO(pubKeyHash []byte) []TXOutput  {
	var UTXOs   []TXOutput
	var db = u.Blockchain.Db

	var err = db.View(func(tx *bolt.Tx) error {
		var b = tx.Bucket([]byte(utxoBucket))
		var c = b.Cursor()

		for k,v:=c.First();k!=nil;k,v=c.Next(){
			var outs = DeserializeOutputs(v)

			for _,out:=range outs.Outputs{
				if out.IsLockedWithKey(pubKeyHash){
					UTXOs = append(UTXOs,out)
				}
			}
		}

		return nil
	})
	utils.LogErr(err)
	return UTXOs
}

func (u UTXOSet)CountTransactions()int  {
	var db = u.Blockchain.Db
	var counter = 0

	var err = db.View(func(tx *bolt.Tx) error {
		var b = tx.Bucket([]byte(utxoBucket))
		var c = b.Cursor()

		for k , _:= c.First();k!=nil;k ,_ =c.Next(){
			counter++
		}
		return nil
	})
	utils.LogErr(err)

	return counter
}

func (u UTXOSet)Reindex()  {
	var db = u.Blockchain.Db
	var bucketName = []byte(utxoBucket)

	var err = db.Update(func(tx *bolt.Tx) error {
		var err = tx.DeleteBucket(bucketName)
		if err !=nil&&err!=bolt.ErrBucketNotFound{
			utils.LogErr(err)
		}

		_,err = tx.CreateBucket(bucketName)
		utils.LogErr(err)
		return nil
	})

	utils.LogErr(err)

	var UTXO = u.Blockchain.FindUTXO()

	err = db.Update(func(tx *bolt.Tx) error {
		var b = tx.Bucket(bucketName)

		for txId,outs:=range UTXO{
			var key , err = hex.DecodeString(txId)
			utils.LogErr(err)

			err = b.Put(key,outs.Serialize())
			utils.LogErr(err)
		}
		return nil
	})
}

func (u UTXOSet)Update(blck *Block){
	var db = u.Blockchain.Db

	var err = db.Update(func(tx *bolt.Tx) error {
		var b = tx.Bucket([]byte(utxoBucket))

		for _,tx:=range blck.Transactions{
			if tx.IsCoinBose() == false{
				for _,vin:=range tx.Vin{
					var updateOuts = TXOutputs{}
					var outsBytes = b.Get(vin.Txid)
					var outs = DeserializeOutputs(outsBytes)

					for outIdx,out:=range outs.Outputs{
						if outIdx !=vin.Vout{
							updateOuts.Outputs = append(updateOuts.Outputs,out)
						}
					}

					if len(updateOuts.Outputs) == 0{
						var err = b.Delete(vin.Txid)
						utils.LogErr(err)
					}else {
						var err = b.Put(vin.Txid,updateOuts.Serialize())
						utils.LogErr(err)
					}
				}
			}
			var newOutputs = TXOutputs{}
			for _,out:=range tx.Vout{
				newOutputs.Outputs = append(newOutputs.Outputs,out)
			}

			var err = b.Put(tx.ID,newOutputs.Serialize())
			utils.LogErr(err)
		}

		return nil
	})
	utils.LogErr(err)
}
















