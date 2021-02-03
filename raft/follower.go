package raft

func Vote(s *server,vr *VoteRequest)  {
	state := Candidate
	vote := true
	if s.log.LastLogTerm > vr.lastLogTerm || s.log.LastLogIndex > vr.lastLogIndex || s.Term() > vr.term{
		state = Follower
		vote = false
	}else if s.state == Candidate || s.votedFor == ""{
		vote = false
	}else if s.state == Leader{
		state = Follower
		vote = false
	}else{
		s.votedFor = vr.name
	}
	vrp := &VoteResponse{
		vote: 			vote,
		name: 			s.name,
		state:          state,
		ip: 			vr.serverIp,
		port: 			vr.serverPort,
	}
	SendVoteResponse(vrp)

	peer := s.peers[vr.name]
	UpdatePeer(peer,peer.Name,peer.IP,peer.Port,state,vr.lastLogIndex,
		vr.lastLogTerm,peer.heartbeatInterval,peer.lastActivity)
}
