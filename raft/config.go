package raft

import (
	"errors"
	"time"
)

const (
	Stopped      = "stopped"
	Initialized  = "initialized"
	Follower     = "follower"
	Candidate    = "candidate"
	Leader       = "leader"
	Snapshotting = "snapshotting"
)

const (
	MaxLogEntriesPerRequest         = 2000
	NumberOfLogEntriesAfterSnapshot = 200
	MaxServerRecLen = 1000
)

const (
	DefaultHeartbeatInterval = 50 * time.Millisecond
	DefaultElectionTimeout = 150 * time.Millisecond
	DefaultMonitorTimeout = 150 * time.Millisecond
)

const ElectionTimeoutThresholdPercent = 0.8


var NotLeaderError = errors.New("raft.Server: Not current leader")
var DuplicatePeerError = errors.New("raft.Server: Duplicate peer")
var CommandTimeoutError = errors.New("raft: Command timeout")
var StopError = errors.New("raft: Has been stopped")

//udp包头部代码块
const(
	AddPeerOrder = "1001"
	VoteOrder = "1002"
	VoteBackOrder = "1003"
	HeartBeatOrder = "1004"
	AppendLogEntryOrder = "1005"
	AppendLogEntryResponseOrder = "1006"
	AddLogEntryOrder = "1007"
	StopServer = "1008"
	DelPeerOrder = "1009"
	JoinRaftOrder = "1010"
	JoinRaftResponseOrder = "1011"
	AddKeyOrder = "1012"
	GetAllPeersOrder = "1013"
	GetAllPeersResponseOrder = "1014"
	GetOneServerOrder = "1015"
	GetOneServerResponseOrder = "1016"
	AddPeerResponseOrder = "1017"
	MonitorRequestOrder = "1018"
	MonitorResponseOrder = "1019"
)

//默认广播的地址和端口
const(
	UdpIp = "192.168.1.255"
	UdpPort = 10001
	InitUdpPort = 10003
)
