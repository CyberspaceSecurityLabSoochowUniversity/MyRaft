package raft

type VoteRequest struct {
	name            string
	term 			uint64
	lastLogIndex   	uint64
	lastLogTerm 	uint64
	ip              string
	port			int
}

func NewVoteRequest()  {

}
