package raft

import (
	client "../socket/client"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"
)

type heartBeat struct {
	Name 		string
	Term 		uint64
	ServerIp  	string
	ServerPort 	int
	LogIndex 	uint64
	LogTerm 	uint64
	Ip 			string
	Port        int
}

func NewHeartBeat(name string,term uint64,serverIp string,serverPort int,
	logIndex uint64,logTerm uint64,ip string,port int) *heartBeat {
	hb := &heartBeat{
		Name: 					name,
		Term: 					term,
		ServerIp: 				serverIp,
		ServerPort: 			serverPort,
		LogIndex: 				logIndex,
		LogTerm: 				logTerm,
		Ip: 					ip,
		Port: 					port,
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
	if hb.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendHeartBeat: IP is blank!")
		return
	}
	if hb.Port <= 0{
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
	client.NewClient(hb.Ip,hb.Port,data)
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
	fmt.Println("LeaderLoop Start")
	s.leader = s.name
	s.SetHeartbeatInterval(DefaultHeartbeatInterval)
	startHeartBeat(s)		//领导者一上线就得广播心跳
	go func() {
		for s.State() == Leader{
			fmt.Println(s.peers)
			data := make([]byte, MaxServerRecLen)
			_, _, err := conn.ReadFromUDP(data)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Server(%s):Read udp content error:%s\n", s.ip, err.Error())
				return
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
			case AddPeerOrder:
				apr := ReceiveAddPeerRequest(data1.Value)
				err = s.AddPeer(apr)
				if err != nil {
					fmt.Fprintln(os.Stdout, "Server add peer error:", err.Error())
				}
				break
			case AddPeerResponseOrder:
				aprp := ReceiveAddPeerResponse(data1.Value)
				_,ok := s.peers[aprp.Name]
				if !ok{
					peer := NewPeer(aprp.Name,aprp.IP,aprp.RecPort,aprp.State,aprp.LastLogIndex,
						aprp.LastLogTerm,aprp.HeartbeatInterval,aprp.LastActivity)
					s.peers[peer.Name] = peer		//添加对等点
				}
				break
			case AppendLogEntryOrder:
				ale := ReceiveAppendLogEntryRequest(data1.Value)
				entry := s.log.entries[ale.LogIndex:]
				alerp := NewAppendLogEntryResponse(s.name,entry,ale.ServerIp,ale.ServerPort)
				SendAppendLogEntryResponse(alerp)
				break
			case VoteOrder:
				vr := ReceiveVoteRequest(data1.Value)
				Vote(s,vr)
				break
			case AddLogEntryOrder:
				addle :=ReceiveAddLogEntryRequest(data1.Value)
				key := addle.Key
				value := addle.Value
				logEntry := NewLogEntry(s.log,0, s.log.LastLogIndex+1, s.currentTerm,key,value)
				s.log.entries = append(s.log.entries, *logEntry)
				s.log.LastLogIndex += 1
				s.log.LastLogTerm = s.Term()
				//将日志条目持久化log文件中

				break
			case StopServer:
				stopRequest := ReceiveStopRequest(data1.Value)
				if stopRequest.Name == s.name{
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
			case GetOneServerOrder:
				gsr := ReceiveGetServerRequest(data1.Value)
				if gsr.Name == s.Name(){
					gsrp := NewGetServerResponse(s,gsr.ClientIp,gsr.ClientPort,gsr.EntranceId,gsr.EntrancePort)
					SendGetServerResponse(gsrp)
				}
				break
			case MonitorRequestOrder:
				mr := ReceiveMonitorRequest(data1.Value)
				mrp := NewMonitorResponse(s.Name(),mr.EntranceIp,mr.EntranceRecPort)
				SendMonitorResponse(mrp)
				break
			}
		}
	}()

	for s.State() == Leader {

	}
}
