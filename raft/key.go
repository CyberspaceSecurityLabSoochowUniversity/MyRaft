package raft

import (
	client "../socket/client"
	"encoding/json"
	"fmt"
	"os"
)

type AddKeyRequest struct {
	EntranceId  string		//客户端决定发送给哪个入口节点
	Key 		string
	Value 		string
	Ip          string
	Port        int
}

func NewAddKeyRequest(entranceId string,key string,value string,ip string,port int) *AddKeyRequest {
	akr := &AddKeyRequest{
		EntranceId: entranceId,
		Key: key,
		Value: value,
		Ip: ip,
		Port: port,
	}
	return akr
}

func SendAddKeyRequest(akr *AddKeyRequest)  {
	if akr.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendAddKeyRequest: IP is blank!")
		return
	}
	if akr.Port <= 0{
		fmt.Fprintln(os.Stdout,"SendAddKeyRequest: Port is incorrect!")
		return
	}
	message,err := json.Marshal(akr)
	d := client.Date{Id: AddKeyOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendAddKeyRequest: Error converting data into Json!")
		return
	}
	client.NewClient(akr.Ip,akr.Port,data)
}

func ReceiveAddKeyRequest(message []byte) *AddKeyRequest {
	akr := new(AddKeyRequest)
	err := json.Unmarshal(message,&akr)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveAddKeyRequest Error:",err.Error())
		return nil
	}
	return akr
}
