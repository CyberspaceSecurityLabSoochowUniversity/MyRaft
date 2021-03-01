package raft

import "os"

//状态机（类似于数据库）
type StateMachine struct {
	file        *os.File
	db          map[string]string
	path        string
}


