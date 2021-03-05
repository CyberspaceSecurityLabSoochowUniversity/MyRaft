package raft

import (
	client "../socket/client"
	"encoding/json"
	"fmt"
	"os"
)

type VoteRequest struct {
	Name            string
	ServerIp	    string					//服务器ip
	ServerPort		int						//服务器接收数据端口
	Term 			uint64
	LastLogIndex   	uint64
	LastLogTerm 	uint64
	Ip              string					//广播地址
	Port			int						//广播端口
}

func NewVoteRequest(s *server,ip string,port int) *VoteRequest {
	if s.Term() < s.log.LastLogTerm{
		fmt.Fprintln(os.Stdout,"VoteRequest:New a vote request error:term < lastLogTerm")
		return nil
	}
	vr := &VoteRequest{
		Name: 					s.name,
		ServerIp: 				s.ip,
		ServerPort: 			s.recPort,
		Term: 					s.Term(),
		LastLogIndex: 			s.log.LastLogIndex,
		LastLogTerm: 			s.log.LastLogTerm,
		Ip: 					ip,
		Port: 					port,
	}
	return vr
}

func SendVoteRequest(vr *VoteRequest)  {
	if vr.Ip == ""{
		fmt.Fprintln(os.Stdout,"VoteRequest: IP is blank!")
		return
	}
	if vr.Port <= 0{
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
	client.NewClient(vr.Ip,vr.Port,data)
}

func ReceiveVoteRequest(message []byte) *VoteRequest {
	vr := new(VoteRequest)
	err := json.Unmarshal(message,&vr)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveVoteRequest Error:",err.Error())
		return nil
	}
	return vr
}

type VoteResponse struct {
	Vote            bool
	Name            string
	State			string
	Ip              string
	Port			int
}

func SendVoteResponse(vrp *VoteResponse)  {
	if vrp.Ip == ""{
		fmt.Fprintln(os.Stdout,"VoteResponse: IP is blank!")
		return
	}
	if vrp.Port <= 0{
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
	client.NewClient(vrp.Ip,vrp.Port,data)
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

func Vote(s *server,vr *VoteRequest)  {
	if s.name == vr.Name{
		return
	}
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if s.VotedTerm() == s.Term(){
		return
	}
	state := Candidate
	vote := true
	//if s.log.LastLogTerm > vr.LastLogTerm || s.log.LastLogIndex > vr.LastLogIndex || s.Term() > vr.Term{
	if s.log.LastLogTerm > vr.LastLogTerm || s.log.LastLogIndex > vr.LastLogIndex{
		state = Follower
		vote = false
	}else if s.state == Candidate{
		vote = false
	}else if s.state == Leader{
		state = Follower
		vote = false
	}else{
		vote = true
	}
	if vote == true{
		s.votedTerm = s.currentTerm
		s.votedFor = vr.Name
		s.currentTerm += 1
	}
	vrp := &VoteResponse{
		Vote: 			vote,
		Name: 			s.name,
		State:          state,
		Ip: 			vr.ServerIp,
		Port: 			vr.ServerPort,
	}

	//peer := s.peers[vr.Name]
	//UpdatePeer(peer,peer.Name,peer.IP,peer.Port,state,vr.LastLogIndex,
	//	vr.LastLogTerm,peer.heartbeatInterval,peer.lastActivity)

	SendVoteResponse(vrp)
}