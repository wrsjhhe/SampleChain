package utils

import (
	"bytes"
	"encoding/binary"
	"log"
)

//将int64转换为字节数组
func IntToHex(num int64) []byte  {

	var buff = new(bytes.Buffer)

	var err = binary.Write(buff,binary.BigEndian,num)
	LogErr(err)

	return buff.Bytes()

}

func LogErr(err error)  {
	if err!=nil{
		log.Panic(err)
	}
}
