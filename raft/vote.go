package raft

import (
	client "../socket/client"
	"encoding/json"
	"fmt"
	"os"
)

type VoteRequest struct {
	name            string
	serverIp	    string					//服务器ip
	serverPort		int						//服务器接收数据端口
	term 			uint64
	lastLogIndex   	uint64
	lastLogTerm 	uint64
	ip              string					//广播地址
	port			int						//广播端口
}

func NewVoteRequest(s *server,ip string,port int) *VoteRequest {
	if s.Term() < s.log.LastLogTerm{
		fmt.Fprintln(os.Stdout,"VoteRequest:New a vote request error:term < lastLogTerm")
		return nil
	}
	vr := &VoteRequest{
		name: 					s.name,
		serverIp: 				s.ip,
		serverPort: 			s.recPort,
		term: 					s.Term(),
		lastLogIndex: 			s.log.LastLogIndex,
		lastLogTerm: 			s.log.LastLogTerm,
		ip: 					ip,
		port: 					port,
	}
	return vr
}

func SendVoteRequest(vr *VoteRequest)  {
	if vr.ip == ""{
		fmt.Fprintln(os.Stdout,"VoteRequest: IP is blank!")
		return
	}
	if vr.port <= 0{
		fmt.Fprintln(os.Stdout,"VoteRequest: Port is incorrect!")
		return
	}
	message,err := json.Marshal(vr)
	d := client.Date{Id: VoteOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"VoteRequest: Error converting data into Json!")
		return
	}
	client.NewClient(vr.ip,vr.port,data)
}

func ReceiveVoteVoteRequest(message []byte) *VoteRequest {
	vr := new(VoteRequest)
	err := json.Unmarshal(message,&vr)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveVoteVoteRequest Error:",err.Error())
		return nil
	}
	return vr
}

type VoteResponse struct {
	vote            bool
	name            string
	state			string
	ip              string
	port			int
}

func SendVoteResponse(vrp *VoteResponse)  {
	if vrp.ip == ""{
		fmt.Fprintln(os.Stdout,"VoteResponse: IP is blank!")
		return
	}
	if vrp.port <= 0{
		fmt.Fprintln(os.Stdout,"VoteResponse: Port is incorrect!")
		return
	}
	message,err := json.Marshal(vrp)
	d := client.Date{Id: VoteBackOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"VoteResponse: Error converting data into Json!")
		return
	}
	client.NewClient(vrp.ip,vrp.port,data)
}

func ReceiveVoteResponse(message []byte) *VoteResponse {
	vrp := new(VoteResponse)
	err := json.Unmarshal(message,&vrp)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveVoteResponse Error:",err.Error())
		return nil
	}
	return vrp
}