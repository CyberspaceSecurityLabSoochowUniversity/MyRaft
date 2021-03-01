package raft

import (
	client "../socket/client"
	"encoding/json"
	"fmt"
	"os"
)

type joinRequest struct {
	name string
	sip   string
	recPort int
	ip string
	port int
}

func NewJoinRequest(name string,serverIp string,serverPort int,ip string,port int) *joinRequest {
	jr := &joinRequest{
		name: name,
		sip: serverIp,
		recPort: serverPort,
		ip: ip,
		port: port,
	}
	return jr
}

func SendJoinRequest(jr *joinRequest)  {
	if jr.ip == ""{
		fmt.Fprintln(os.Stdout,"SendJoinRequest: IP is blank!")
		return
	}
	if jr.port <= 0{
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
	client.NewClient(jr.ip,jr.port,data)
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
	join   bool
	name   string
	ip     string
	port   int
}

func NewJoinResponse(join bool,name string,ip string,port int) *joinResponse {
	jrp := &joinResponse{
		join: join,
		name: name,
		ip: ip,
		port: port,
	}
	return jrp
}

func SendJoinResponse(jrp *joinResponse)  {
	if jrp.ip == ""{
		fmt.Fprintln(os.Stdout,"SendJoinResponse: IP is blank!")
		return
	}
	if jrp.port <= 0{
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
	client.NewClient(jrp.ip,jrp.port,data)
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