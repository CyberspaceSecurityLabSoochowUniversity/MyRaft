package raft

import (
	client "../socket/client"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Peer struct {
	Name              string
	IP  			  string
	Port			  int				//接收消息的Port
	state             string
	LastLogIndex      uint64
	LastLogTerm       uint64
	heartbeatInterval time.Duration
	lastActivity      time.Time
}

func NewPeer(name string,ip string,recPort int,state string,index uint64,term uint64,heartbeatInterval time.Duration,
	lastActivity time.Time) *Peer {
	return &Peer{
		Name:              name,
		IP:  			   ip,
		Port:              recPort,
		state: state,
		LastLogIndex: index,
		LastLogTerm: term,
		heartbeatInterval: heartbeatInterval,
		lastActivity: lastActivity,
	}
}

//func (p *Peer) clone() *Peer {
//	return &Peer{
//		Name:              p.Name,
//		IP:  			   p.IP,
//		Port:              p.Port,
//		state:			   p.state,
//		LastLogIndex: 	   p.LastLogIndex,
//		LastLogTerm:       p.LastLogIndex,
//		heartbeatInterval: p.heartbeatInterval,
//		lastActivity:      p.lastActivity,
//	}
//}

func (p *Peer) setHeartbeatInterval(duration time.Duration) {
	p.heartbeatInterval = duration
}



type AddPeerRequest struct {
	Name              string
	IP  			  string
	Port			  int				//接收消息的Port
	state             string
	LastLogIndex      uint64
	LastLogTerm       uint64
	heartbeatInterval time.Duration
	lastActivity      time.Time

	ip    string		//广播的地址
	port  int			//广播的端口
}

func NewAddPeerRequest(server *server,ip string,port int) *AddPeerRequest {
	apr := &AddPeerRequest{
		Name: 					server.name,
		IP: 					server.ip,
		Port: 					server.recPort,
		state: 					server.state,
		LastLogIndex: 			server.log.LastLogIndex,
		LastLogTerm: 			server.log.LastLogTerm,
		heartbeatInterval: 		server.heartbeatInterval,
		lastActivity: 			time.Now(),
		ip: 					ip,
		port:					port,
	}
	return apr
}

func SendAddPeerRequest(apr *AddPeerRequest) {
	if apr.ip == ""{
		fmt.Fprintln(os.Stdout,"SendAddPeerRequest: IP is blank!")
		return
	}
	if apr.port <= 0{
		fmt.Fprintln(os.Stdout,"SendAddPeerRequest: Port is incorrect!")
		return
	}
	message,err := json.Marshal(apr)
	d := client.Date{Id: AddPeerOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendAddPeerRequest: Error converting data into Json!")
		return
	}
	client.NewClient(apr.ip,apr.port,data)
}

func ReceiveAddPeerRequest(message []byte) *AddPeerRequest {
	apr := new(AddPeerRequest)
	err := json.Unmarshal(message,&apr)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveAddPeerRequest Error:",err.Error())
		return nil
	}
	return apr
}

func UpdatePeer(peer *Peer,name string,ip string,recPort int,state string,index uint64,
	term uint64,heartbeatInterval time.Duration,lastActivity time.Time)  {
	peer.Name = name
	peer.IP = ip
	peer.Port = recPort
	peer.state = state
	peer.LastLogIndex = index
	peer.LastLogTerm = term
	peer.heartbeatInterval = heartbeatInterval
	peer.lastActivity = lastActivity
}


type DelPeerRequest struct {
	Name              string
	ip    			  string
	port  			  int
}

func NewDelPeerRequest(name string, ip string, port int) *DelPeerRequest {
	dpr := &DelPeerRequest{
		Name: 	name,
		ip: 	ip,
		port:	port,
	}
	return dpr
}

func SendDelPeerRequest(dpr *DelPeerRequest) {
	if dpr.ip == ""{
		fmt.Fprintln(os.Stdout,"SendDelPeerRequest: IP is blank!")
		return
	}
	if dpr.port <= 0{
		fmt.Fprintln(os.Stdout,"SendDelPeerRequest: Port is incorrect!")
		return
	}
	message,err := json.Marshal(dpr)
	d := client.Date{Id: DelPeerOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendDelPeerRequest: Error converting data into Json!")
		return
	}
	client.NewClient(dpr.ip,dpr.port,data)
}

func ReceiveDelPeerRequest(message []byte) *DelPeerRequest {
	dpr := new(DelPeerRequest)
	err := json.Unmarshal(message,&dpr)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveDelPeerRequest Error:",err.Error())
		return nil
	}
	return dpr
}