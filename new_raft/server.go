package new_raft

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
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

type Server struct{
	IP     string
	port   uint64
	name   string
	path   string
	state  string
	currentTerm uint64
	waitTime   uint64
	votedfor   string
	leader     string
	log  *Log
	heartbeattime uint64
	routineGroup  sync.WaitGroup
	mutex sync.RWMutex
	electionTimeout   time.Duration
}

func newServer(IP string,port uint64,name string,path string)(*Server,error){
	if IP==""{
		return nil, errors.New("Raft.Server:IP cannot be blank")
	}
	s:=&Server{//创建服务器
		IP:      IP,
		port:    port,
		name:    name,
		path:    path,
		state:   Stopped,  //开始工作状态是停止的
	}
	return s,nil  //返回创建的服务器
}

func (s *Server)setState(state string){
	s.mutex.Lock()
	defer s.mutex.Unlock()

	prevState:=s.state
	prevLeader:=s.leader

	s.state=state
	if state==Leader{//如果服务器是领导者需要存储领导者名称和建立连接的对等点
		s.leader=s.IP
	}
}

func (s *Server)Term() uint64{
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.currentTerm

}

func (s *Server)WaitTime() uint64{
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.waitTime

}

func (s *Server)updateCurrentTerm(term uint64,leadername string){

}

func (s *Server)ElectionTimeout() time.Duration{
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.electionTimeout
}

func (s *Server) promotable() bool {
	return s.log.currentIndex() > 0
}

func (s *Server)Followerloop(){
	s.heartbeattime=5
	//节点状态是跟随者时，端口不断监听，判断收到的消息类型
	var (
		conn   *net.UDPConn
		err    error
		//reader *bufio.Reader
	)

	if conn, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 3000,
	}); err != nil {
		panic(err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			panic(err)
		}
	}()

	//选举超时定时器
	timer:=time.NewTicker(s.ElectionTimeout())


	for {
		update:=false
		data := make([]byte, 1024)
		n,_,err:=conn.ReadFromUDP(data)
		if err!=nil{
			fmt.Printf("err: %s", err.Error())
			continue
		}

		c:=make(chan []byte,1024)
		if data!=nil{
			c<-data
		}

		select{
			case <-c:
				//判断收到的消息类型：
			//1、领导者的心跳（Nopcommand），定时器重置
			//2、领导者请求添加日志（AppendEntriesRequest）
			//3、候选者请求投票（RequestVoteRequest）
			msg_str := strings.Split(string(data[0:n]), "|")
			switch msg_str[1] {
				case "heartbeat"://领导者发送的心跳
					s.heartbeattime=0
					fmt.Println("after:",s.heartbeattime)
				case "RequestvoteRequest"://候选者请求投票
					term,_:=strconv.ParseUint(msg_str[3],10,64)
					logindex,_:=strconv.ParseUint(msg_str[4],10,64)
					waittime,_:=strconv.ParseUint(msg_str[5],10,64)
					update:=s.processRequestVoteRequest(msg_str[0],term,logindex,waittime)
				case "AppendEntriesRequest"://领导者请求添加日志
				}
		case <-timer.C:
			//超时未收到来自领导者的心跳，则认为未收到集群中领导者下线
			if s.promotable() {
				s.setState(Candidate)
			} else {
				update = true		//由于日志条目为空则无法提升为候选者
			}
		}

		if update{
			timer=time.NewTicker(s.ElectionTimeout())
		}
	}
}

func (s *Server)Leaderloop(){
	//端口不断监听
	var (
		conn   *net.UDPConn
		err    error
		//reader *bufio.Reader
	)

	if conn, err = net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 3000,
	}); err != nil {
		panic(err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			panic(err)
		}
	}()

	s.currentTerm=0

	//广播心跳
	go s.heartbeat()
}

func (s *Server)heartbeat(){

	// 这里设置发送者的IP地址，自己查看一下自己的IP自行设定
	laddr := net.UDPAddr{
		IP:   net.ParseIP(s.IP),
		Port: 3000,
	}

	// 这里设置接收者的IP地址为广播地址
	raddr := net.UDPAddr{
		IP:   net.IPv4(255, 255, 255, 255),
		Port: 3000,
	}
	conn, err := net.DialUDP("udp", &laddr, &raddr)
	if err != nil {
		println(err.Error())
		return
	}

	for{//每隔一段时间发送一次心跳
		lastIndex, lastTerm := s.log.lastInfo()
		line:="AppendEntriesRequest|"+strconv.Itoa(time.Now().Nanosecond())+"|"+strconv.Itoa(int(lastTerm))+"|"+strconv.Itoa(int(lastIndex))+"|Nop"
		_, err = conn.Write([]byte(s.IP+line))
		time.Sleep(50*time.Millisecond)
	}

}

func (s *Server)processRequestVoteRequest(candidatename string,term uint64,logindex uint64,waitTime uint64)  bool { //处理投票请求并返回响应给候选者
	var (
		conn net.Conn
		err  error
	)

	if conn, err = net.Dial("udp", candidatename); err != nil {
		panic(err)
	}

	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Println(err.Error())
		}
	}()

	lastIndex, lastTerm := s.log.lastInfo()
	if term<s.Term(){ //当前服务器任期大于请求者任期，不投票
		line:=s.IP+"|RequestVote|"+strconv.Itoa(time.Now().Nanosecond())+"|"+strconv.Itoa(int(s.Term()))+"|"+strconv.Itoa(int(lastIndex))+"|"+strconv.Itoa(int(s.WaitTime()))+"|2|false|"+candidatename
		_,err=conn.Write([]byte(line))
		return false
	}
	if term>s.Term(){ //当前服务器任期小于请求者任期，任期需要改变
		s.updateCurrentTerm(term," ")
	}else if s.votedfor!=" "&&s.votedfor!=candidatename{
		line:=s.IP+"|RequestVote|"+strconv.Itoa(time.Now().Nanosecond())+"|"+strconv.Itoa(int(s.Term()))+"|"+strconv.Itoa(int(lastIndex))+"|"+strconv.Itoa(int(s.WaitTime()))+"|5|false|"+candidatename
		_,err=conn.Write([]byte(line))
		return false
	}
	if lastIndex>logindex||lastTerm>term{
		line:=s.IP+"|RequestVote|"+strconv.Itoa(time.Now().Nanosecond())+"|"+strconv.Itoa(int(s.Term()))+"|"+strconv.Itoa(int(lastIndex))+"|"+strconv.Itoa(int(s.WaitTime()))+"|3|false|"+candidatename
		_,err=conn.Write([]byte(line))
		return false
	}

	if s.WaitTime()>waitTime{
		line:=s.IP+"|RequestVote|"+strconv.Itoa(time.Now().Nanosecond())+"|"+strconv.Itoa(int(s.Term()))+"|"+strconv.Itoa(int(lastIndex))+"|"+strconv.Itoa(int(s.WaitTime()))+"|4|false|"+candidatename
		_,err=conn.Write([]byte(line))
		return false
	}
	s.votedfor=candidatename  //可以投票，投给请求的候选者
	line:=s.IP+"|RequestVote|"+strconv.Itoa(time.Now().Nanosecond())+"|"+strconv.Itoa(int(s.Term()))+"|"+strconv.Itoa(int(lastIndex))+"|"+strconv.Itoa(int(s.WaitTime()))+"|1|true|"+candidatename
	_,err=conn.Write([]byte(line))
	return true
}