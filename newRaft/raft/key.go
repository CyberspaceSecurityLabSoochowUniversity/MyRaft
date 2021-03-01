package raft

import (
	client "../socket/client"
	"encoding/json"
	"fmt"
	"os"
)

type AddKeyRequest struct {
	entranceId  string		//客户端决定发送给哪个入口节点
	key 		string
	value 		string
	ip          string
	port        int
}

func NewAddKeyRequest(entranceId string,key string,value string,ip string,port int) *AddKeyRequest {
	akr := &AddKeyRequest{
		entranceId: entranceId,
		key: key,
		value: value,
		ip: ip,
		port: port,
	}
	return akr
}

func SendAddKeyRequest(akr *AddKeyRequest)  {
	if akr.ip == ""{
		fmt.Fprintln(os.Stdout,"SendAddKeyRequest: IP is blank!")
		return
	}
	if akr.port <= 0{
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
	client.NewClient(akr.ip,akr.port,data)
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
