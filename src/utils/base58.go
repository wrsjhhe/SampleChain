package utils

import (
	"math/big"
	"bytes"
)

var b58Alphabet = []byte("123456789ABCDEFGHIJKLMNPQRSTUVWXYZabcdefghijklmnpqrstuvwxyz")

//将字节数组转化为base58
func Base58Encode(input []byte)[]byte  {
	var result []byte

	var x = big.NewInt(0).SetBytes(input)

	var base = big.NewInt(int64(len(b58Alphabet)))
	var zero = big.NewInt(0)
	var mod = &big.Int{}

	for x.Cmp(zero)!=0{
		x.DivMod(x,base,mod)
		result = append(result,b58Alphabet[mod.Int64()])
	}

	ReverseBytes(result)
	for b:= range input{
		if b == 0x00{
			result = append([]byte{b58Alphabet[0]},result...)
		}else {
			break
		}
	}
	return result
}

//将base58转化为字节数组
func Base58Decode(input []byte)[]byte  {
	var result = big.NewInt(0)
	var zeroBytes = 0

	for b:=range input{
		if b == 0x00{
			zeroBytes++
		}
	}

	var payload = input[zeroBytes:]
	for _,b := range payload{
		var charIndex = bytes.IndexByte(b58Alphabet,b)
		result.Mul(result,big.NewInt(58))
		result.Add(result,big.NewInt(int64(charIndex)))
	}

	var decoded = result.Bytes()
	decoded = append(bytes.Repeat([]byte{byte(0x00)},zeroBytes),decoded...)

	return decoded
}
