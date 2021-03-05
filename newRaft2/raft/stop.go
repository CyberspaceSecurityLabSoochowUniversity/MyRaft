package raft

import (
	client "../socket/client"
	"encoding/json"
	"fmt"
	"os"
	"time"
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
	if err != nil {
		fmt.Fprintln(os.Stdout, "SendStopRequest: Error converting data into Json!")
		return
	}

	go func() {
		for i:=0;i<3;i++{
			client.NewClient(stopRequest.Ip,stopRequest.Port,data)
			time.Sleep(50*time.Millisecond)
		}
	}()
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

type MonitorRequest struct {
	EntranceIp string
	EntranceRecPort int
	Ip string
	Port int
}

func NewMonitorRequest(entranceIp string,entrancePort int,ip string,port int) *MonitorRequest {
	mr := &MonitorRequest{
		EntranceIp: entranceIp,
		EntranceRecPort: entrancePort,
		Ip: ip,
		Port: port,
	}
	return mr
}

func SendMonitorRequest(mr *MonitorRequest)  {
	if mr.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendMonitorRequest: IP is blank!")
		return
	}
	if mr.Port <= 0{
		fmt.Fprintln(os.Stdout,"SendMonitorRequest: Port is incorrect!")
		return
	}
	message,err := json.Marshal(mr)
	d := client.Date{Id: MonitorRequestOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendMonitorRequest: Error converting data into Json!")
		return
	}

	go func() {
		for i:=0;i<3;i++{
			client.NewClient(mr.Ip,mr.Port,data)
			time.Sleep(150*time.Millisecond)
		}
	}()
}

func ReceiveMonitorRequest(message []byte) *MonitorRequest {
	mr := new(MonitorRequest)
	err := json.Unmarshal(message,&mr)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveMonitorRequest Error:",err.Error())
		return nil
	}
	return mr
}

type MonitorResponse struct {
	Name string
	Ip string
	Port int
}

func NewMonitorResponse(name string,ip string,port int) *MonitorResponse {
	mrp := &MonitorResponse{
		Name: name,
		Ip: ip,
		Port: port,
	}
	return mrp
}

func SendMonitorResponse(mrp *MonitorResponse)  {
	if mrp.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendMonitorResponse: IP is blank!")
		return
	}
	if mrp.Port <= 0{
		fmt.Fprintln(os.Stdout,"SendMonitorResponse: Port is incorrect!")
		return
	}
	message,err := json.Marshal(mrp)
	d := client.Date{Id: MonitorResponseOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendMonitorResponse: Error converting data into Json!")
		return
	}

	//go func() {
	//	for i:=0;i<3;i++{
	//		client.NewClient(mrp.Ip,mrp.Port,data)
	//		time.Sleep(10*time.Millisecond)
	//	}
	//}()

	client.NewClient(mrp.Ip,mrp.Port,data)
}

func ReceiveMonitorResponse(message []byte) *MonitorResponse {
	mrp := new(MonitorResponse)
	err := json.Unmarshal(message,&mrp)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveMonitorResponse Error:",err.Error())
		return nil
	}
	return mrp
}