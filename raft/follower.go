package raft

import (
	client "../socket/client"
	"encoding/json"
	"fmt"
	"net"
	"os"
)



func followerLoop(s *server,conn *net.UDPConn) {
	timeoutChan := afterBetween(s.ElectionTimeout(), s.ElectionTimeout()*2)		//开启选举超时
	for s.State() == Follower {
		update := false
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
		case VoteOrder:
			vr := ReceiveVoteVoteRequest(data1.Value)
			Vote(s,vr)
			break
		case HeartBeatOrder:
			hb := ReceiveHeartBeat(data1.Value)
			s.currentTerm = hb.term
			s.leader = hb.name
			if hb.logIndex > s.log.LastLogIndex{
				//向领导者发送一个添加日志请求
				ale := NewAppendLogEntry(s.name,s.ip,s.recPort,s.log.LastLogIndex,hb.serverIp,hb.serverPort)
				SendAppendLogEntryRequest(ale)
			}

			//此处是否需要考虑领导者节点日志索引小于其他一些节点的情况？如果考虑则需要让所有节点与领导者保持一致

			timeoutChan = afterBetween(s.ElectionTimeout(), s.ElectionTimeout()*2)	//重置选举超时
			break
		case AppendLogEntryResponseOrder:
			alerp := ReceiveAppendLogEntryResponse(data1.Value)
			for _,i := range alerp.entry{
				s.log.entries = append(s.log.entries, i)
			}
			s.log.LastLogIndex += uint64(len(alerp.entry))
			s.log.LastLogTerm = uint64(alerp.entry[len(alerp.entry)-1].Term)
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

		select {
		//发送添加对等点请求

		case <-timeoutChan:
			if s.promotable(){
				s.SetState(Candidate)
				s.currentTerm += 1
				return
			} else {
				update = true		//由于日志条目为空则无法提升为候选者
			}
			break
		}
		if update {			//未提升，则产生新一轮的跟随者事件
			timeoutChan = afterBetween(s.ElectionTimeout(), s.ElectionTimeout()*2)
		}
	}
}
