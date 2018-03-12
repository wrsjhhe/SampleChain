package block

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"utils"
	"crypto/sha256"
	"bytes"
	"golang.org/x/crypto/ripemd160"
)

const version = byte(0x00)
const walletFile = "wallet.dat"
const addressChecksumLen = 4

//钱包储存私钥和公钥
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}


//创建一个钱包
func NewWallet()*Wallet  {
	var private,public = newKeyPair()
	var wallet = Wallet{private,public}

	return &wallet
}


//返回钱包地址
func (w Wallet)GetAddress()[]byte  {
	var pubKeyHash = HashPubKey(w.PublicKey)

	var versionPayload = append([]byte{version},pubKeyHash...)
	var checksum = checksum(versionPayload)

	var fullPayload = append(versionPayload,checksum...)
	var address = utils.Base58Encode(fullPayload)

	return address
}

//使用RIPEMD160(SHA256(PubKey))算法，对公钥进行哈希
func HashPubKey(pubKey []byte) []byte {
	var publicSHA256 = sha256.Sum256(pubKey)


	var RIPEMD160Hasher = ripemd160.New()
	var _,err = RIPEMD160Hasher.Write(publicSHA256[:])
	utils.LogErr(err)

	var publicRIPEMD160 = RIPEMD160Hasher.Sum(nil)
	return publicRIPEMD160
}

//检查地址是否有效
func ValidateAddress(address string)bool  {
	var pubKeyHash = utils.Base58Decode([]byte(address))
	var actualChecksum = pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	var version = pubKeyHash[0]

	pubKeyHash = pubKeyHash[1:len(pubKeyHash)-addressChecksumLen]
	var targetChecksum = checksum(append([]byte{version},pubKeyHash...))

	var b = bytes.Compare(actualChecksum,targetChecksum)
	b++
	return bytes.Compare(actualChecksum,targetChecksum) == 0
}

//将HashPubKey的结果加上地址生成算法版本的前缀再使用SHA256(SHA256(payload))哈希，计算校验和
//校验和的结果是哈希的前四个字节
func checksum(payload []byte)[]byte{
	var firstSHA = sha256.Sum256(payload)
	var secondSHA = sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

//基于椭圆曲线算法ECDSA生成秘钥对
func newKeyPair()(ecdsa.PrivateKey,[]byte)  {
	var curve = elliptic.P256()
	var private,err = ecdsa.GenerateKey(curve,rand.Reader)
	utils.LogErr(err)
	var pubKey = append(private.PublicKey.X.Bytes(),private.PublicKey.Y.Bytes()...)

	return *private,pubKey
}
