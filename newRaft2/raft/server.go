package raft

import (
	client "../socket/client"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"strconv"
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
	votedTerm  uint64					//投赞成票的任期
	vote       int					    //投赞成的票数
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
	stopHeartBeatChan chan bool				//停止心跳通道
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
	VotedTerm() uint64

	Log() *Log
	LogPath() string
	CommitIndex() uint64
	IsLogEmpty() bool
	LogEntries() []LogEntry

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


	Init(ip string,port int) error
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

func (s *server) VoteFor() string {
	return s.votedFor
}

func (s *server) VotedTerm() uint64 {
	return s.votedTerm
}

func (s *server) MaxWaitTimeServer() string {
	return s.maxWaitTimeServer
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

func (s *server) LogEntries() []LogEntry {
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
		return fmt.Errorf("peer already in raft")
	}
	if s.name != request.Name {
		peer := NewPeer(request.Name,request.IP,request.RecPort,request.State,request.LastLogIndex,
			request.LastLogTerm,request.HeartbeatInterval,request.LastActivity)
		s.peers[peer.Name] = peer		//添加对等点
		apr1 := NewAddPeerResponse(s,request.IP,request.Port)
		SendAddPeerResponse(apr1)
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

func (s *server) Init(ip string,port int) error {
	if s.Running() {		//服务器正在运行则返回错误
		return fmt.Errorf("raft.Server: Server already running[%v]", s.state)
	}

	if ip == ""{
		return fmt.Errorf("raft.Server: Entrance id is blank")
	}

	if port <= 0{
		return fmt.Errorf("raft.Server: Entrance port is incorrect")
	}

	address := s.ip + ":" + strconv.Itoa(InitUdpPort)
	addr,err := net.ResolveUDPAddr("udp",address)
	if err != nil{
		fmt.Fprintf(os.Stderr,"Server(%s):New a upd server error:%s\n",address,err.Error())
		return err
	}
	conn,err := net.ListenUDP("udp",addr)
	if err != nil{
		fmt.Fprintf(os.Stderr,"Server(%s):New a listenupd error:%s\n",address,err.Error())
		return err
	}
	defer conn.Close()

	jr := NewJoinRequest(s.name,s.ip,InitUdpPort,ip,port)
	SendJoinRequest(jr)

	start := false
	for start != true{
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
			return err
		}

		switch data1.Id {
		case JoinRaftResponseOrder:
			jrp := ReceiveJoinResponse(data1.Value)
			if jrp.Join == true{
				start = true
			}else{
				return fmt.Errorf("raft.Server: Entrance not allowed to join the raft")
			}
			break
		}
	}
	s.log.LastLogIndex = 1
	s.state = Initialized		//设置服务器状态为已初始化过了
	return nil
}

func (s *server) loop(conn *net.UDPConn)  {

	state := s.State()

	for state != Stopped {
		switch state {		//四种状态对应四种处理方式
		case Follower:
			followerLoop(s,conn)
			break
		case Candidate:
			candidateLoop(s,conn)
			break
		case Leader:
			leaderLoop(s,conn)
			break
		//case Snapshotting:
		//	snapshotLoop(s,conn)
		}
		state = s.State()	//处理完了可能需要改变服务器状态
	}
}

func (s *server) Start() error {
	if s.Running() {
		return fmt.Errorf("raft.Server: Server already running[%v]", s.state)
	}
	if s.State() != Initialized {
		return fmt.Errorf("raft.Server: Server is not initialized")
	}
	s.SetState(Follower)		//S设置服务器状态为跟随者

	//此处代码为：启动服务器监听，一旦接收到值则运行相应的代码块
	address := s.ip + ":" + strconv.Itoa(s.recPort)
	addr,err := net.ResolveUDPAddr("udp",address)
	if err != nil{
		fmt.Fprintf(os.Stderr,"Server(%s):New a upd server error:%s\n",address,err.Error())
		return err
	}
	conn,err := net.ListenUDP("udp",addr)
	if err != nil{
		fmt.Fprintf(os.Stderr,"Server(%s):New a listenupd error:%s\n",address,err.Error())
		return err
	}
	defer conn.Close()

	apr := NewAddPeerRequest(s,UdpIp,UdpPort)
	SendAddPeerRequest(apr)

	s.loop(conn)

	return nil
}

func (s *server) Stop() {
	if s.State() == Stopped {		//已关闭过了
		return
	}

	//删除对等点
	dpr := NewDelPeerRequest(s.name,UdpIp,UdpPort)
	SendDelPeerRequest(dpr)

	// make sure all goroutines have stopped before we close the log
	//关闭之前确定所有的日志协程都被关闭了
	//s.routineGroup.Wait()

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

func (s *server) promotable() bool {
	return s.log.LastLogIndex > 0
}

type GetServerRequest struct {
	EntranceId string
	Name string
	EntranceIp string
	EntrancePort int
	ClientIp string
	ClientPort int
	Ip string
	Port int
}

func NewGetServerRequest(entranceId string,clientIp string,clientPort int,name string,
	ip string,port int) *GetServerRequest {
	gsr := &GetServerRequest{
		EntranceId: entranceId,
		Name: name,
		ClientIp: clientIp,
		ClientPort: clientPort,
		Ip: ip,
		Port: port,
	}
	return gsr
}

func SendGetServerRequest(gsr * GetServerRequest)  {
	if gsr.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendGetServerRequest: IP is blank!")
		return
	}
	if gsr.Port <= 0{
		fmt.Fprintln(os.Stdout,"SendGetServerRequest: Port is incorrect!")
		return
	}
	message,err := json.Marshal(gsr)
	d := client.Date{Id: GetOneServerOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendGetServerRequest: Error converting data into Json!")
		return
	}
	for i:=0;i<3;i++{
		client.NewClient(gsr.Ip,gsr.Port,data)
		time.Sleep(19*time.Millisecond)
	}

}
func ReceiveGetServerRequest(message []byte) *GetServerRequest {
	gsr := new(GetServerRequest)
	err := json.Unmarshal(message,&gsr)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveGetServerRequest Error:",err.Error())
		return nil
	}
	return gsr
}

type GetServerResponse struct {
	Name string
	Path string
	ServerIp string
	State string
	RecPort int
	Context string
	CurrentTerm uint64
	VoteTerm uint64
	VoteFor string
	Leader  string
	ElectionTimeout time.Duration
	HeartBeatInterval time.Duration
	MaxLogEntriesPerRequest uint64
	ClientIp string
	ClientPort int
	Ip string
	Port int
}

func NewGetServerResponse(s *server,clientIp string,clientPort int,ip string,port int) *GetServerResponse {
	gsrp := &GetServerResponse{
		Name: s.Name(),
		Path: s.Path(),
		ServerIp: s.Ip(),
		State: s.State(),
		RecPort: s.RecPort(),
		Context: s.Context(),
		CurrentTerm: s.Term(),
		VoteTerm: s.VotedTerm(),
		VoteFor: s.VoteFor(),
		Leader: s.Leader(),
		ElectionTimeout: s.ElectionTimeout(),
		HeartBeatInterval: s.HeartbeatInterval(),
		MaxLogEntriesPerRequest: s.maxLogEntriesPerRequest,
		ClientIp: clientIp,
		ClientPort: clientPort,
		Ip: ip,
		Port: port,
	}
	return gsrp
}

func SendGetServerResponse(gsrp *GetServerResponse)  {
	if gsrp.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendGetServerResponse: IP is blank!")
		return
	}
	if gsrp.Port <= 0{
		fmt.Fprintln(os.Stdout,"SendGetServerResponse: Port is incorrect!")
		return
	}
	message,err := json.Marshal(gsrp)
	d := client.Date{Id: GetOneServerResponseOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendGetServerResponse: Error converting data into Json!")
		return
	}
	for i:=0;i<3;i++{
		client.NewClient(gsrp.Ip,gsrp.Port,data)
		time.Sleep(19*time.Millisecond)
	}

}
func ReceiveGetServerResponse(message []byte) *GetServerResponse {
	gsrp := new(GetServerResponse)
	err := json.Unmarshal(message,&gsrp)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveGetServerResponse Error:",err.Error())
		return nil
	}
	return gsrp
}
