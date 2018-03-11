package block

import (
	"fmt"
	"os"
	"io/ioutil"
	"utils"
	"encoding/gob"
	"crypto/elliptic"
	"bytes"
)

//储存钱包集合
type Wallets struct {
	Wallets map[string]*Wallet
}

//从一个存在的文件中读取钱包的集合
func NewWallets()(*Wallets,error)  {
	var wallets = Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	var err = wallets.LoadFromFile()

	return &wallets,err
}

//生成一个新的钱包并返回该钱包地址
func (ws *Wallets)CreateWallet() string {
	var wallet = NewWallet()
	var address = fmt.Sprintf("%s",wallet.GetAddress())

	ws.Wallets[address] = wallet

	return address
}

//返回钱包集合中的地址数组
func (ws *Wallets)GetAddresses()[]string  {
	var addresses []string

	for address:=range ws.Wallets{
		addresses = append(addresses,address)
	}
	return addresses
}


//通过地址返回钱包
func (ws Wallets)GetWallet(address string) Wallet  {
	return *ws.Wallets[address]
}

//从文件中读取钱包集合
func (ws Wallets)LoadFromFile()error  {
	if _,err :=os.Stat(walletFile);os.IsNotExist(err){
		return err
	}

	fileContent,err:=ioutil.ReadFile(walletFile)
	utils.LogErr(err)

	var wallets Wallets
	gob.Register(elliptic.P256())
	var decoder = gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	utils.LogErr(err)

	ws.Wallets = wallets.Wallets
	return err
}

//保存钱包到文件中
func (ws Wallets)SaveToFile()  {
	var content bytes.Buffer

	gob.Register(elliptic.P256())

	var encoder = gob.NewEncoder(&content)
	var err = encoder.Encode(ws)
	utils.LogErr(err)

	err = ioutil.WriteFile(walletFile,content.Bytes(),0644)
	utils.LogErr(err)
}



