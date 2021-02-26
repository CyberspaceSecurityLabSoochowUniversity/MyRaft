package raft

import (
	"errors"
	"sync"
)

//入口节点（元服务器）
type entrance struct {
	id              string              //集群的标识
	ip 				string				//ip
	recPort 		int					//接收udp消息地址
	mutex      		sync.RWMutex		//读写锁
	context     	string				//详细信息
	currentLeader   string				//当前集群领导者
	currentTerm     uint64				//当前集群任期
	peer            map[string]string   //集群节点的信息（名称：ip）
	pLen            uint64				//集群大小（节点数量）
}

type Entrance interface {
	Id()    			string
	Ip()    			string
	SetIp(string)
	RecPort() 			int
	Context() 			string
	CurrentLeader() 	string
	CurrentTerm()   	uint64
	Peer()    			map[string]string
	PeerLen() 			uint64
	Start()
}

func NewEntrance(id string,ip string,port int,ct string) (error,Entrance) {
	if id == ""{
		return errors.New("entrance id is blank"),nil
	}
	if ip == ""{
		return errors.New("entrance ip is blank"),nil
	}
	if port <= 0 {
		return errors.New("entrance port is incorrect"),nil
	}
	et := &entrance{
		id: id,
		ip: ip,
		recPort: port,
		context: ct,
	}
	return nil,et
}

func (et *entrance) Id() string {
	return et.id
}

func (et *entrance) Ip() string {
	et.mutex.RLock()
	defer et.mutex.RUnlock()
	return et.ip
}

func (et *entrance) SetIp(ip string)  {
	if ip != "" && ip != et.ip{
		et.ip = ip
	}
}

func (et *entrance) RecPort() int {
	return et.recPort
}

func (et *entrance) Context() string {
	return et.context
}

func (et *entrance) CurrentLeader() string {
	et.mutex.RLock()
	defer et.mutex.RUnlock()
	return et.currentLeader
}

func (et *entrance) CurrentTerm() uint64 {
	et.mutex.RLock()
	defer et.mutex.RUnlock()
	return et.currentTerm
}

func (et *entrance) Peer() map[string]string {
	return et.peer
}

func (et *entrance) PeerLen() uint64 {
	return uint64(len(et.peer))
}

func (et *entrance) Start() {

}






