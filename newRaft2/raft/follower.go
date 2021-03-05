package raft

import (
	client "../socket/client"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
)

func followerLoop(s *server,conn *net.UDPConn) {
	fmt.Println("FollowerLoop Start")
	fmt.Println(s.peers)
	timeoutChan := afterBetween(s.ElectionTimeout(), s.ElectionTimeout()*2)		//开启选举超时
	//timeoutChan := time.NewTicker(afterBetween1(s.ElectionTimeout(), s.ElectionTimeout()*2))

	go func() {
		//fmt.Println("444444")
		for s.State() == Follower{
			data := make([]byte, MaxServerRecLen)
			_, _, err := conn.ReadFromUDP(data)
			if s.State() != Follower{
				break
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "Server(%s):Read udp content error:%s\n", s.ip, err.Error())
				continue
			}
			data = bytes.Trim(data,"\x00")

			//这里需要根据接收内容类型进行相应处理
			data1 := new(client.Date)
			err = json.Unmarshal(data, &data1)
			fmt.Println(data1.Id)
			if err != nil {
				fmt.Fprintln(os.Stdout, "ReceiveData Error:", err.Error())
				continue
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
				vr := ReceiveVoteRequest(data1.Value)
				Vote(s,vr)
				break
			case HeartBeatOrder:
				hb := ReceiveHeartBeat(data1.Value)
				fmt.Println("Follower: Receive heatBeat:",hb)
				s.currentTerm = hb.Term
				s.leader = hb.Name
				if hb.LogIndex > s.log.LastLogIndex{
					//向领导者发送一个添加日志请求
					ale := NewAppendLogEntry(s.name,s.ip,s.recPort,s.log.LastLogIndex,hb.ServerIp,hb.ServerPort)
					SendAppendLogEntryRequest(ale)
				}

				////此处是否需要考虑领导者节点日志索引小于其他一些节点的情况？如果考虑则需要让所有节点与领导者保持一致
				timeoutChan = afterBetween(s.ElectionTimeout(), s.ElectionTimeout()*2)
				break
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
				fmt.Println("del")
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
				fmt.Println(mr)
				mrp := NewMonitorResponse(s.Name(),mr.EntranceIp,mr.EntranceRecPort)
				SendMonitorResponse(mrp)
				break
			default:
				break
			}
		}
	}()


	for s.State() == Follower {
		update := false
		select {
		case <-timeoutChan:
			fmt.Println("Follower: TimeoutChan Get Value")
			if s.promotable(){
				s.currentTerm += 1
				s.SetState(Candidate)
				return
			} else {
				update = true		//由于日志条目为空则无法提升为候选者
			}
			break
		default:
			break
		}
		if update {			//未提升，则产生新一轮的跟随者事件
			timeoutChan = afterBetween(s.ElectionTimeout(), s.ElectionTimeout()*2)
		}
	}
}
