package raft

import (
	"errors"
	"fmt"
	"path"
	"sync"
	"time"
)

type server struct {
	//*eventDispatcher

	name        string					//名称
	path        string					//物理地址
	ip 			string					//ip地址
	state       string					//状态
	sendPort	int						//发送的端口
	recPort		int						//接收的端口

	context     string					//详细介绍
	currentTerm uint64					//当前任期
	currentWaitTime uint64				//当前等待值
	maxWaitTimeServer string			//最大等待值服务器

	votedFor   string					//投票的服务器名称
	log        *Log						//日志信息
	leader     string					//领导者名称
	peers      map[string]*Peer			//对等点
	mutex      sync.RWMutex				//读写锁
	//syncedPeer map[string]bool

	//stopped           chan bool
	//c                 chan *ev
	electionTimeout   time.Duration		//选举超时时间
	heartbeatInterval time.Duration		//心跳时间

	//snapshot *Snapshot					//快照

	stateMachine            *StateMachine	//状态机
	maxLogEntriesPerRequest uint64			//最大一次请求的日志条目个数

	routineGroup sync.WaitGroup				//并发队列锁
}

type Server interface {
	Name() string
	Path() string
	Ip() string
	State() string
	SendPort() int
	RecPort() int
	Context() string
	Term() uint64
	WaitTime() uint64
	MaxWaitTimeServer() string
	VoteFor()	string

	Log() *Log
	LogPath() string
	CommitIndex() uint64
	IsLogEmpty() bool
	LogEntries() []*LogEntry

	Leader() string

	Peers() map[string]*Peer
	AddPeer(request *AddPeerRequest) error
	RemovePeer(name string) error

	ElectionTimeout() time.Duration
	SetElectionTimeout(duration time.Duration)
	HeartbeatInterval() time.Duration
	SetHeartbeatInterval(duration time.Duration)

	StateMachine() *StateMachine

	SnapshotPath(lastIndex uint64, lastTerm uint64) string

	MemberCount() int
	QuorumSize() int

	GetState() string

	//AppendEntries(req *AppendEntriesRequest) *AppendEntriesResponse
	//RequestVote(req *RequestVoteRequest) *RequestVoteResponse
	//RequestSnapshot(req *SnapshotRequest) *SnapshotResponse
	//SnapshotRecoveryRequest(req *SnapshotRecoveryRequest) *SnapshotRecoveryResponse


	Init() error
	Start() error
	Stop()
	Running() bool
	//Do(command Command) (interface{}, error)
	//TakeSnapshot() error
	//LoadSnapshot() error
	//AddEventListener(string, EventListener)
	//FlushCommitIndex()


	SetState(state string)
}

func NewServer(name string, path string, context string, ip string, sendPort int, recPort int,stateMachine *StateMachine) (Server, error) {
	if name == ""{
		return nil, errors.New("raft.Server: Name cannot be blank")
	}
	if path == ""{
		return nil, errors.New("raft.Server: Path cannot be blank")
	}
	if ip == ""{
		return nil, errors.New("raft.Server: IP cannot be blank")
	}
	if sendPort == 0{
		return nil, errors.New("raft.Server: SendPort cannot be zero")
	}
	if recPort == 0{
		return nil, errors.New("raft.Server: RecPort cannot be zero")
	}
	s := &server{		//创建服务器
		name:                    name,
		path:                    path,
		ip:						 ip,
		state:                   Stopped,		//开始状态为停止工作
		sendPort: 				 sendPort,
		recPort: 				 recPort,
		context:				 context,
		log:                     NewLog(path),
		peers:                   make(map[string]*Peer),
		electionTimeout:         DefaultElectionTimeout,		//默认的选举超时时间
		heartbeatInterval:       DefaultHeartbeatInterval,		//默认的心跳间隔时间
		stateMachine:            stateMachine,
		maxLogEntriesPerRequest: MaxLogEntriesPerRequest,		//最大的日志条目请求数
	}
	//s.eventDispatcher = newEventDispatcher(s)		//服务器创建一个新的事件调度器
	return s, nil

}

func (s *server) Name() string {
	return s.name
}

func (s *server) Path() string {
	return s.path
}

func (s *server) Ip() string {
	return s.ip
}

func (s *server) State() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state
}

func (s *server) SendPort() int {
	return s.sendPort
}

func (s *server) RecPort() int {
	return s.recPort
}

func (s *server) Context() string {
	return s.context
}

func (s *server) Term() uint64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.currentTerm
}

func (s *server) WaitTime() uint64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.currentWaitTime
}

func (s *server) MaxWaitTimeServer() string {
	return s.maxWaitTimeServer
}

func (s *server) VoteFor() string {
	return s.votedFor
}

func (s *server) Log() *Log {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.log
}

func (s *server) LogPath() string {
	return path.Join(s.path,"log")
}

func (s *server) CommitIndex() uint64 {
	s.log.mutex.RLock()
	defer s.log.mutex.RUnlock()
	return s.log.commitIndex
}

func (s *server) IsLogEmpty() bool {
	return s.log.isEmpty()
}

func (s *server) LogEntries() []*LogEntry {
	s.log.mutex.RLock()
	defer s.log.mutex.RUnlock()
	return s.log.entries
}

func (s *server) Leader() string {
	return s.leader
}

func (s *server) Peers() map[string]*Peer {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.peers
	//peers := make(map[string]*Peer)
	//for name, peer := range s.peers {
	//	peers[name] = peer.clone()
	//}
	//return peers
}

//func NewPeer(name string,ip string,recPort int,state string,index uint64,term uint64,heartbeatInterval time.Duration,
//	lastActivity time.Time) *Peer {

func (s *server) AddPeer(request *AddPeerRequest) error {
	if s.peers[request.Name] != nil {
		return nil
	}
	if s.name != request.Name {
		peer := NewPeer(request.Name,request.ip,request.port,request.state,request.LastLogIndex,
			request.LastLogTerm,request.heartbeatInterval,request.lastActivity)

		//if s.State() == Leader {		//如果服务器都是领导者了，那对等点就需要开始心跳了
		//	peer.startHeartbeat()
		//}

		s.peers[peer.Name] = peer		//添加对等点
	}
	return nil
}

func (s *server) RemovePeer(name string) error {

	//移除自己！想什么呢，做不到
	if name != s.Name() {
		peer := s.peers[name]
		if peer == nil {
			return fmt.Errorf("raft Server: Peer not found: %s", name)
		}

		//if s.State() == Leader {
		//	s.routineGroup.Add(1)
		//	go func() {
		//		defer s.routineGroup.Done()
		//		peer.stopHeartbeat(true)
		//	}()
		//}

		delete(s.peers, name)		//删除
	}

	return nil
}

func (s *server) ElectionTimeout() time.Duration {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.electionTimeout
}

func (s *server) SetElectionTimeout(duration time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.electionTimeout = duration
}

func (s *server) HeartbeatInterval() time.Duration {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.heartbeatInterval
}

func (s *server) SetHeartbeatInterval(duration time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.heartbeatInterval = duration
	for _, peer := range s.peers {
		peer.setHeartbeatInterval(duration)		//对每一个对等点都进行更新心跳时间间隔
	}
}

func (s *server) StateMachine() *StateMachine {
	return s.stateMachine
}

func (s *server) SnapshotPath(lastIndex uint64, lastTerm uint64) string {
	return path.Join(s.path, "snapshot", fmt.Sprintf("%v_%v.ss", lastTerm, lastIndex))
}

func (s *server) MemberCount() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.peers) + 1		//+1代表的是服务器自己
}

func (s *server) QuorumSize() int {
	return (s.MemberCount() / 2) + 1
}

func (s *server) GetState() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return fmt.Sprintf("Name: %s, State: %s, Term: %v, CommitedIndex: %v ", s.name, s.state, s.currentTerm, s.log.commitIndex)
}

func (s *server) Init() error {
	if s.Running() {		//服务器正在运行则返回错误
		return fmt.Errorf("raft.Server: Server already running[%v]", s.state)
	}

	s.state = Initialized		//设置服务器状态为已初始化过了
	return nil
}

func (s *server) Start() error {
	if s.Running() {
		return fmt.Errorf("raft.Server: Server already running[%v]", s.state)
	}
	if err := s.Init(); err != nil {
		return err
	}
	s.SetState(Follower)		//S设置服务器状态为跟随者

	//此处代码为：启动服务器监听，一旦接收到值则运行相应的代码块

	return nil
}

func (s *server) Stop() {
	if s.State() == Stopped {		//已关闭过了
		return
	}

	// make sure all goroutines have stopped before we close the log
	//关闭之前确定所有的日志协程都被关闭了
	s.routineGroup.Wait()

	//s.log.close()		//日志关闭

	s.SetState(Stopped)		//设置服务器状态为已关闭
}

func (s *server) Running() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return (s.state != Stopped) && (s.state != Initialized)
}

func (s *server) SetState(state string) {		//测试专用，不能随便用，会造成集群停机
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	s.state = state
}

