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

//逆序一个字节数组
func ReverseBytes(data []byte){
	for i,j :=0,len(data)-1;i<j;i,j = i+1,j-1{
		data[i],data[j] = data[j],data[i]
	}
}

func LogErr(err error)  {
	if err!=nil{
		log.Panic(err)
	}
}
