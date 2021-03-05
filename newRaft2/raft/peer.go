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
	RecPort			  int				//接收消息的Port
	State             string
	LastLogIndex      uint64
	LastLogTerm       uint64
	HeartbeatInterval time.Duration
	LastActivity      time.Time

	Ip    string		//广播的地址
	Port  int			//广播的端口
}

func NewAddPeerRequest(server *server,ip string,port int) *AddPeerRequest {
	apr := &AddPeerRequest{
		Name: 					server.name,
		IP: 					server.ip,
		RecPort: 				server.recPort,
		State: 					server.state,
		LastLogIndex: 			server.log.LastLogIndex,
		LastLogTerm: 			server.log.LastLogTerm,
		HeartbeatInterval: 		server.heartbeatInterval,
		LastActivity: 			time.Now(),
		Ip: 					ip,
		Port:					port,
	}
	return apr
}

func SendAddPeerRequest(apr *AddPeerRequest) {
	if apr.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendAddPeerRequest: IP is blank!")
		return
	}
	if apr.Port <= 0{
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
	for i:=0;i<3;i++{
		client.NewClient(apr.Ip,apr.Port,data)
		time.Sleep(15*time.Millisecond)
	}

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

type AddPeerResponse struct {
	Name              string
	IP  			  string
	RecPort			  int				//接收消息的Port
	State             string
	LastLogIndex      uint64
	LastLogTerm       uint64
	HeartbeatInterval time.Duration
	LastActivity      time.Time

	Ip    string		//广播的地址
	Port  int			//广播的端口
}

func NewAddPeerResponse(server *server,ip string,port int) *AddPeerResponse {
	aprp := &AddPeerResponse{
		Name: 					server.name,
		IP: 					server.ip,
		RecPort: 				server.recPort,
		State: 					server.state,
		LastLogIndex: 			server.log.LastLogIndex,
		LastLogTerm: 			server.log.LastLogTerm,
		HeartbeatInterval: 		server.heartbeatInterval,
		LastActivity: 			time.Now(),
		Ip: 					ip,
		Port:					port,
	}
	return aprp
}

func SendAddPeerResponse(aprp *AddPeerResponse) {
	if aprp.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendAddPeerResponse: IP is blank!")
		return
	}
	if aprp.Port <= 0{
		fmt.Fprintln(os.Stdout,"SendAddPeerResponse: Port is incorrect!")
		return
	}
	message,err := json.Marshal(aprp)
	d := client.Date{Id: AddPeerResponseOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendAddPeerResponse: Error converting data into Json!")
		return
	}

	for i:=0;i<3;i++{
		client.NewClient(aprp.Ip,aprp.Port,data)
		time.Sleep(20*time.Millisecond)
	}

}

func ReceiveAddPeerResponse(message []byte) *AddPeerResponse {
	aprp := new(AddPeerResponse)
	err := json.Unmarshal(message,&aprp)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveAddPeerResponse Error:",err.Error())
		return nil
	}
	return aprp
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
	Ip    			  string
	Port  			  int
}

func NewDelPeerRequest(name string, ip string, port int) *DelPeerRequest {
	dpr := &DelPeerRequest{
		Name: 	name,
		Ip: 	ip,
		Port:	port,
	}
	return dpr
}

func SendDelPeerRequest(dpr *DelPeerRequest) {
	if dpr.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendDelPeerRequest: IP is blank!")
		return
	}
	if dpr.Port <= 0{
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
	for i:=0;i<3;i++{
		client.NewClient(dpr.Ip,dpr.Port,data)
		time.Sleep(17*time.Millisecond)
	}

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

type GetAllPeersRequest struct {
	EntranceId  string
	ClientIp    string
	ClientPort  int
	Ip          string
	Port        int
}

func NewGetAllPeersRequest(entranceId string,clientIp string,clientPort int,ip string,port int) *GetAllPeersRequest {
	gapr := &GetAllPeersRequest{
		EntranceId: entranceId,
		ClientIp: clientIp,
		ClientPort: clientPort,
		Ip: ip,
		Port: port,
	}
	return gapr
}

func SendGetAllPeersRequest(gapr *GetAllPeersRequest)  {
	if gapr.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendGetAllPeersRequest: IP is blank!")
		return
	}
	if gapr.Port <= 0{
		fmt.Fprintln(os.Stdout,"SendGetAllPeersRequest: Port is incorrect!")
		return
	}
	message,err := json.Marshal(gapr)
	d := client.Date{Id: GetAllPeersOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendGetAllPeersRequest: Error converting data into Json!")
		return
	}
	for i:=0;i<3;i++{
		client.NewClient(gapr.Ip,gapr.Port,data)
		time.Sleep(20*time.Millisecond)
	}

}

func ReceiveGetAllPeersRequest(message []byte) *GetAllPeersRequest {
	gapr := new(GetAllPeersRequest)
	err := json.Unmarshal(message,&gapr)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveGetAllPeersRequest Error:",err.Error())
		return nil
	}
	return gapr
}

type GetAllPeersResponse struct {
	EntranceId string
	Peers map[string]string
	Ip string
	Port int
}

func NewGetAllPeersResponse(entranceId string,peers map[string]string,ip string,port int) *GetAllPeersResponse {
	gaprp := &GetAllPeersResponse{
		EntranceId: entranceId,
		Peers: peers,
		Ip: ip,
		Port: port,
	}
	return gaprp
}

func SendGetAllPeersResponse(gaprp *GetAllPeersResponse)  {
	if gaprp.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendGetAllPeersResponse: IP is blank!")
		return
	}
	if gaprp.Port <= 0{
		fmt.Fprintln(os.Stdout,"SendGetAllPeersResponse: Port is incorrect!")
		return
	}
	message,err := json.Marshal(gaprp)
	d := client.Date{Id: GetAllPeersResponseOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendGetAllPeersResponse: Error converting data into Json!")
		return
	}
	for i:=0;i<3;i++{
		client.NewClient(gaprp.Ip,gaprp.Port,data)
		time.Sleep(25*time.Millisecond)
	}

}

func ReceiveGetAllPeersResponse(message []byte) *GetAllPeersResponse {
	gaprp := new(GetAllPeersResponse)
	err := json.Unmarshal(message,&gaprp)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveGetAllPeersResponse Error:",err.Error())
		return nil
	}
	return gaprp
}