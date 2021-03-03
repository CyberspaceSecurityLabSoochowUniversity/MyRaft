package raft

import (
	client "../socket/client"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
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
	sign            uint64				//存储添加值的次数，同时作为添加的日志条目的唯一标识
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
	Sign()              uint64
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
		peer: make(map[string]string),
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

func (et *entrance) Sign() uint64 {
	et.mutex.RLock()
	defer et.mutex.RUnlock()
	return et.sign
}

func (et *entrance) Start() {

	address := et.ip + ":" + strconv.Itoa(et.recPort)
	addr,err := net.ResolveUDPAddr("udp",address)
	if err != nil{
		fmt.Fprintf(os.Stderr,"Entrance(%s):New a upd server error:%s\n",address,err.Error())
		return
	}
	conn,err := net.ListenUDP("udp",addr)
	if err != nil{
		fmt.Fprintf(os.Stderr,"Entrance(%s):New a listenupd error:%s\n",address,err.Error())
		return
	}
	defer conn.Close()

	for{
		data := make([]byte, MaxServerRecLen)
		_, _, err := conn.ReadFromUDP(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Entrance(%s):Read udp content error:%s\n", et.ip, err.Error())
			continue
		}
		data = bytes.Trim(data,"\x00")

		//这里需要根据接收内容类型进行相应处理
		data1 := new(client.Date)
		err = json.Unmarshal(data, &data1)
		if err != nil {
			fmt.Fprintln(os.Stdout, "ReceiveData Error:", err.Error())
			return
		}

		switch data1.Id {
		case JoinRaftOrder:
			jr := ReceiveJoinRequest(data1.Value)
			var ok bool			//判断名称和ip是否已经存在，存在则为true
			for key,value := range et.peer{
				if key == jr.Name || value == jr.Ip{
					ok = true
				}
			}
			go func() {
				result := false
				if !ok {
					var a string
					fmt.Printf("是否让%s 加入集群(y/n)",jr.Name+":"+jr.Ip)
					fmt.Scanf("%s",&a)
					if a == "y"{
						et.peer[jr.Name] = jr.Ip
						result = true
					}
				}else{
					fmt.Fprintln(os.Stdout,"Server already in Raft")
				}

				jrp := NewJoinResponse(result,jr.Name,jr.Sip,jr.RecPort)
				SendJoinResponse(jrp)
			}()
			break

		case HeartBeatOrder:
			hb := ReceiveHeartBeat(data1.Value)
			et.currentTerm = hb.Term
			et.currentLeader = hb.Name

			break

		case AddKeyOrder:
			//客户端发来的添加key/value的请求，这里选择单独发送给leader，后期如果测试丢包则采用广播形式
			akr := ReceiveAddKeyRequest(data1.Value)
			if akr.EntranceId == et.Id(){
				et.sign += 1
				addle := NewAddLogEntry(et.sign,akr.Key,akr.Value,et.peer[et.currentLeader],UdpPort)
				SendAddLogEntryRequest(addle)
			}
			break

		case StopServer:
			stopRequest := ReceiveStopRequest(data1.Value)
			if stopRequest.EntranceId == et.Id(){
				_,ok := et.peer[stopRequest.Name]
				if ok{
					stopRequest.Ip = et.peer[stopRequest.Name]
					stopRequest.Port = UdpPort
					SendStopRequest(stopRequest)
				}
			}
			break

		case DelPeerOrder:
			dpr := ReceiveDelPeerRequest(data1.Value)
			_,ok := et.peer[dpr.Name]
			if ok == true{
				delete(et.peer,dpr.Name)
			}
			break

		case GetAllPeersOrder:
			gapr := ReceiveGetAllPeersRequest(data1.Value)
			if gapr.EntranceId == et.Id(){
				gaprp := NewGetAllPeersResponse(et.Id(),et.peer,gapr.ClientIp,gapr.ClientPort)
				SendGetAllPeersResponse(gaprp)
			}
			break

		case GetOneServerOrder:
			gsr := ReceiveGetServerRequest(data1.Value)
			if gsr.EntranceId == et.Id(){
				gsr.EntranceId = et.Ip()
				gsr.EntrancePort = et.RecPort()
				gsr.Ip = et.peer[gsr.Name]
				gsr.Port = UdpPort
				SendGetServerRequest(gsr)
			}
			break
		case GetOneServerResponseOrder:
			gsrp := ReceiveGetServerResponse(data1.Value)
			gsrp.Ip =gsrp.ClientIp
			gsrp.Port = gsrp.ClientPort
			SendGetServerResponse(gsrp)
			break
		}
	}

}






