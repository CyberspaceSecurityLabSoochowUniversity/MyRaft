package raft

import (
	client "../socket/client"
	"encoding/json"
	"fmt"
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
