package raft

import (
	client "../socket/client"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type joinRequest struct {
	Name string
	Sip   string
	RecPort int
	Ip string
	Port int
}

func NewJoinRequest(name string,serverIp string,serverPort int,ip string,port int) *joinRequest {
	jr := &joinRequest{
		Name: name,
		Sip: serverIp,
		RecPort: serverPort,
		Ip: ip,
		Port: port,
	}
	return jr
}

func SendJoinRequest(jr *joinRequest)  {
	if jr.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendJoinRequest: IP is blank!")
		return
	}
	if jr.Port <= 0{
		fmt.Fprintln(os.Stdout,"SendJoinRequest: Port is incorrect!")
		return
	}
	message,err := json.Marshal(jr)
	d := client.Date{Id: JoinRaftOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendJoinRequest: Error converting data into Json!")
		return
	}

	client.NewClient(jr.Ip,jr.Port,data)
}

func ReceiveJoinRequest(message []byte) *joinRequest {
	jr := new(joinRequest)
	err := json.Unmarshal(message,&jr)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveJoinRequest Error:",err.Error())
		return nil
	}
	return jr
}

type joinResponse struct {
	Join   bool
	Name   string
	Ip     string
	Port   int
}

func NewJoinResponse(join bool,name string,ip string,port int) *joinResponse {
	jrp := &joinResponse{
		Join: join,
		Name: name,
		Ip: ip,
		Port: port,
	}
	return jrp
}

func SendJoinResponse(jrp *joinResponse)  {
	if jrp.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendJoinResponse: IP is blank!")
		return
	}
	if jrp.Port <= 0{
		fmt.Fprintln(os.Stdout,"SendJoinResponse: Port is incorrect!")
		return
	}
	message,err := json.Marshal(jrp)
	d := client.Date{Id: JoinRaftResponseOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendJoinResponse: Error converting data into Json!")
		return
	}
	for i:=0;i<3;i++{
		client.NewClient(jrp.Ip,jrp.Port,data)
		time.Sleep(25*time.Millisecond)
	}

}

func ReceiveJoinResponse(message []byte) *joinResponse {
	jrp := new(joinResponse)
	err := json.Unmarshal(message,&jrp)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveJoinResponse Error:",err.Error())
		return nil
	}
	return jrp
}