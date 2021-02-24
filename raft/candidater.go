package raft

import (
	client "../socket/client"
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
		//循环发起发起投票请求，以防有些节点出现丢包情况和有新节点加入集群的情况
		vr := NewVoteRequest(s,UdpIp,UdpPort)
		SendVoteRequest(vr)

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
		case VoteBackOrder:
			if s.state == Candidate{
				vrp := ReceiveVoteResponse(data1.Value)
				if vrp.vote == true{
					s.vote += 1
					if s.vote >= s.QuorumSize(){
						s.SetState(Leader)
						return
					}
				}else{
					if vrp.state != Candidate{
						s.state = vrp.state
						return
					}
				}
			}
			break
		case HeartBeatOrder:
			hb := ReceiveHeartBeat(data1.Value)
			if hb.logIndex > s.log.LastLogIndex{
				ale := NewAppendLogEntry(s.name,s.ip,s.recPort,s.log.LastLogIndex,hb.serverIp,hb.serverPort)
				SendAppendLogEntryRequest(ale)
			}

			//此处是否需要考虑领导者节点日志索引小于其他一些节点的情况？如果考虑则需要让所有节点与领导者保持一致

			s.currentTerm = hb.term
			s.leader = hb.name
			s.SetState(Follower)
			return
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
		}

		select {
		case <-timeoutChan:			//投票超时则将状态设为Follower
			s.SetState(Follower)
			return
		}
	}
}
