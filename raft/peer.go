package raft

import (
	client "../socket/client"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type Peer struct {
	server            *server
	Name              string
	IP  			  string
	Port			  int
	state             string
	LogIndex      uint64
	//stopChan          chan bool
	heartbeatInterval time.Duration
	lastActivity      time.Time
	sync.RWMutex
}

func NewPeer(server *server, name string, ip string, heartbeatInterval time.Duration) *Peer {
	return &Peer{
		server:            server,
		Name:              name,
		IP:  			   ip,
		Port:              server.recPort,
		heartbeatInterval: heartbeatInterval,
	}
}

func (p *Peer) clone() *Peer {
	p.Lock()
	defer p.Unlock()
	return &Peer{
		server:            p.server,
		Name:              p.Name,
		IP:  			   p.IP,
		Port:              p.Port,
		state:			   p.state,
		LogIndex: 		   p.LogIndex,
		heartbeatInterval: p.heartbeatInterval,
		lastActivity:      p.lastActivity,
	}
}

func (p *Peer) setHeartbeatInterval(duration time.Duration) {
	p.heartbeatInterval = duration
}



type AddPeerRequest struct {
	peer  Peer			//添加的服务器信息
	ip    string		//广播的地址
	port  int			//广播的端口
}

func SendAddPeerRequest(peer *AddPeerRequest) {
	if peer.ip == ""{
		fmt.Fprintln(os.Stdout,"SendAddPeerRequest: IP is blank!")
		return
	}
	if peer.port <= 0{
		fmt.Fprintln(os.Stdout,"SendAddPeerRequest: Port is incorrect!")
		return
	}
	message,err := json.Marshal(&peer)
	d := client.Date{Id: AddPeerOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendAddPeerRequest: Error converting data into Json!")
		return
	}
	client.NewClient(peer.ip,peer.port,data)
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
