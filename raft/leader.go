package raft

import (
	client "../socket/client"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

type heartBeat struct {
	name 		string
	term 		uint64
	serverIp  	string
	serverPort 	int
	logIndex 	uint64
	logTerm 	uint64
	ip 			string
	port        int
}

func NewHeartBeat(name string,term uint64,serverIp string,serverPort int,
	logIndex uint64,logTerm uint64,ip string,port int) *heartBeat {
	hb := &heartBeat{
		name: 					name,
		term: 					term,
		serverIp: 				serverIp,
		serverPort: 			serverPort,
		logIndex: 				logIndex,
		logTerm: 				logTerm,
		ip: 					ip,
		port: 					port,
	}
	return hb
}

func startHeartBeat(s *server)  {
	hb := NewHeartBeat(s.name,s.Term(),s.ip,s.recPort,s.log.LastLogIndex,s.log.LastLogTerm,UdpIp,UdpPort)
	SendHeartBeat(hb)
	s.routineGroup.Add(1)
	go func() {
		defer s.routineGroup.Done()
		heartBeatFunc(s)
	}()
}

func heartBeatFunc(s *server) {
	ticker := time.Tick(s.heartbeatInterval)
	for {		//不断发送心跳
		select {
		case <-s.stopHeartBeatChan:
			return
		case <-ticker:		//一个心跳时间间隔到了
			hb := NewHeartBeat(s.name,s.Term(),s.ip,s.recPort,s.log.LastLogIndex,s.log.LastLogTerm,UdpIp,UdpPort)
			SendHeartBeat(hb)
		}
	}
}


func SendHeartBeat(hb *heartBeat)  {
	if hb.ip == ""{
		fmt.Fprintln(os.Stdout,"SendHeartBeat: IP is blank!")
		return
	}
	if hb.port <= 0{
		fmt.Fprintln(os.Stdout,"SendHeartBeat: Port is incorrect!")
		return
	}
	message,err := json.Marshal(hb)
	d := client.Date{Id: HeartBeatOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendHeartBeat: Error converting data into Json!")
		return
	}
	client.NewClient(hb.ip,hb.port,data)
}

func ReceiveHeartBeat(message []byte) *heartBeat {
	hb := new(heartBeat)
	err := json.Unmarshal(message,&hb)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveHeartBeat Error:",err.Error())
		return nil
	}
	return hb
}

func leaderLoop(s *server,conn *net.UDPConn) {
	s.SetHeartbeatInterval(DefaultHeartbeatInterval)
	startHeartBeat(s)		//领导者一上线就得广播心跳
	for s.State() == Leader {
		data := make([]byte, MaxServerRecLen)
		_, _, err := conn.ReadFromUDP(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Server(%s):Read udp content error:%s\n", s.ip, err.Error())
			continue
		}

		//这里需要根据接收内容类型进行相应处理
		data1 := new(client.Date)
		err = json.Unmarshal(data, &data1)
		if err != nil {
			fmt.Fprintln(os.Stdout, "ReceiveData Error:", err.Error())
			return
		}
		switch data1.Id {
		case AddPeerOrder:
			apr := ReceiveAddPeerRequest(data1.Value)
			err = s.AddPeer(apr)
			if err != nil {
				fmt.Fprintln(os.Stdout, "Server add peer error:", err.Error())
			} else {
				fmt.Fprintln(os.Stdout, "Server add peer success:")
			}
			apr1 := NewAddPeerRequest(s,apr.IP,apr.Port)
			SendAddPeerRequest(apr1)
			break
		case AppendLogEntryOrder:
			ale := ReceiveAppendLogEntryRequest(data1.Value)
			entry := s.log.entries[ale.logIndex:]
			alerp := NewAppendLogEntryResponse(s.name,entry,ale.serverIp,ale.serverPort)
			SendAppendLogEntryResponse(alerp)
			break
		case VoteOrder:
			vr := ReceiveVoteVoteRequest(data1.Value)
			Vote(s,vr)
			break
		case AddLogEntryOrder:
			addle :=ReceiveAddLogEntryRequest(data1.Value)
			key := addle.Key
			value := addle.Value
			logEntry := NewLogEntry(s.log,0, s.log.LastLogIndex+1, s.currentTerm,key,value)
			s.log.entries = append(s.log.entries, *logEntry)

			//将日志条目持久化log文件中

			break
		case StopServer:
			stopRequest := ReceiveStopRequest(data1.Value)
			if stopRequest.name == s.name{
				s.Stop()
				return
			}
			break
		case DelPeerOrder:
			dpr := ReceiveDelPeerRequest(data1.Value)
			_,ok := s.peers[dpr.Name]
			if ok == true{
				delete(s.peers,dpr.Name)
			}
			break

		}
	}
}
