package raft

import (
	client "../socket/client"
	"encoding/json"
	"fmt"
	"os"
)

type StopRequest struct {
	EntranceId 	string
	Name    	string
	Ip   		string
	Port 		int
}

func NewStopRequest(entranceId string,name string,ip string,port int) *StopRequest {
	stopRequest := &StopRequest{
		EntranceId: entranceId,
		Name: name,
		Ip:   ip,
		Port: port,
	}
	return stopRequest
}

func SendStopRequest(stopRequest *StopRequest)  {
	if stopRequest.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendStopRequest: IP is blank!")
		return
	}
	if stopRequest.Port <= 0{
		fmt.Fprintln(os.Stdout,"SendStopRequest: Port is incorrect!")
		return
	}
	message,err := json.Marshal(stopRequest)
	d := client.Date{Id: StopServer,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendStopRequest: Error converting data into Json!")
		return
	}
	client.NewClient(stopRequest.Ip,stopRequest.Port,data)
}

func ReceiveStopRequest(message []byte) *StopRequest {
	stopRequest := new(StopRequest)
	err := json.Unmarshal(message,&stopRequest)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveStopRequest Error:",err.Error())
		return nil
	}
	return stopRequest
}

