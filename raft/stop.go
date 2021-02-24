package raft

import (
	client "../socket/client"
	"encoding/json"
	"fmt"
	"os"
)

type StopRequest struct {
	name    string
	ip   	string
	port 	int
}

func NewStopRequest(name string,ip string,port int) *StopRequest {
	stopRequest := &StopRequest{
		name: name,
		ip:   ip,
		port: port,
	}
	return stopRequest
}

func SendStopRequest(stopRequest *StopRequest)  {
	if stopRequest.ip == ""{
		fmt.Fprintln(os.Stdout,"SendStopRequest: IP is blank!")
		return
	}
	if stopRequest.port <= 0{
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
	client.NewClient(stopRequest.ip,stopRequest.port,data)
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

