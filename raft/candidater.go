package raft

import (
	client "../socket/client"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

func candidateLoop(s *server,conn *net.UDPConn) {
	s.vote = 1	//自己的一票
	timeoutChan := afterBetween(s.ElectionTimeout(), s.ElectionTimeout()*2)
	//此处需要定义投票超时时间
	for s.State() == Candidate {
		if s.vote >= s.QuorumSize(){
			s.SetState(Leader)
			return
		}
		//循环发起发起投票请求，以防有些节点出现丢包情况和有新节点加入集群的情况
		vr := NewVoteRequest(s,UdpIp,UdpPort)
		SendVoteRequest(vr)

		data := make([]byte, MaxServerRecLen)
		_, _, err := conn.ReadFromUDP(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Server(%s):Read udp content error:%s\n", s.ip, err.Error())
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
		case VoteOrder:
			vr := ReceiveVoteVoteRequest(data1.Value)
			Vote(s,vr)
			break
		case VoteBackOrder:
			if s.state == Candidate{
				vrp := ReceiveVoteResponse(data1.Value)
				if vrp.Vote == true{
					s.vote += 1
					//if s.vote >= s.QuorumSize(){
					//	s.SetState(Leader)
					//	return
					//}
				}else{
					if vrp.State != Candidate{
						s.state = vrp.State
						return
					}
				}
			}
			break
		case HeartBeatOrder:
			hb := ReceiveHeartBeat(data1.Value)
			if hb.LogIndex > s.log.LastLogIndex{
				ale := NewAppendLogEntry(s.name,s.ip,s.recPort,s.log.LastLogIndex,hb.ServerIp,hb.ServerPort)
				SendAppendLogEntryRequest(ale)
			}

			//此处是否需要考虑领导者节点日志索引小于其他一些节点的情况？如果考虑则需要让所有节点与领导者保持一致

			s.currentTerm = hb.Term
			s.leader = hb.Name
			s.SetState(Follower)
			return
		case AppendLogEntryResponseOrder:
			alerp := ReceiveAppendLogEntryResponse(data1.Value)
			for _,i := range alerp.Entry{
				s.log.entries = append(s.log.entries, i)
			}
			s.log.LastLogIndex += uint64(len(alerp.Entry))
			s.log.LastLogTerm = uint64(alerp.Entry[len(alerp.Entry)-1].Term)
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
		}

		select {
		case <-timeoutChan:			//投票超时则将状态设为Follower
			s.SetState(Follower)
			return
		}
	}
}
