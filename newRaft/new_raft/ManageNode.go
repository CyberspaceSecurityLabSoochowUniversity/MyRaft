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

type ManageNode struct{
	IP        string
	port      uint64
	//Nodes     map[string]*Server
	Nodes     []string
	leader    string
	electionTimeout   time.Duration
	mutex      sync.RWMutex
}

type manageNode interface {
	FirstNode(NodeIP string) (uint64,bool)
	NewNode(NodeIP string) (uint64,bool)
	OldNode(NodeIP string) (uint64,bool)
}

func newManageNode(IP string,port uint64)(*ManageNode,error){
	if IP==""{
		return nil, errors.New("Raft.Server:IP cannot be blank")
	}
	M:=&ManageNode{//创建服务器
		IP:      IP,
		port:    port,
		Nodes:   nil,
		leader:  nil,
	}
	return M,nil  //返回创建的服务器
}

func (EntryNode *ManageNode)ElectionTimeout() time.Duration{
	EntryNode.mutex.RLock()
	defer EntryNode.mutex.RUnlock()
	return EntryNode.electionTimeout
}

func (EntryNode *ManageNode)FirstNode(NodeIP string) (uint64,bool){//集群中无节点
	EntryNode.Nodes=append(EntryNode.Nodes,NodeIP)
	return 1,true
}

func (EntryNode *ManageNode)NewNode(NodeIP string) (uint64,bool){//如果是新节点且不是集群中第一个节点
	if EntryNode.leader==" "{//集群中有节点但是领导者下线，集群正在选举领导者不允许新节点加入集群
		return 2,false
	}else{////集群中有节点且领导者在线，允许新节点加入集群
		EntryNode.Nodes=append(EntryNode.Nodes,NodeIP)
		return 2,true
	}
}

func (EntryNode *ManageNode)OldNode(NodeIP string) (uint64,bool){//以前的节点重新上线
	return 3,true
}

func (EntryNode *ManageNode)JoinCluster(msg_str []string) (uint64,bool){
	if EntryNode.Nodes==nil{
		return EntryNode.FirstNode(msg_str[0])
	}
	for _,node:= range EntryNode.Nodes{
		if node==msg_str[0]{
			return EntryNode.OldNode(msg_str[0])
		}
		return EntryNode.NewNode(msg_str[0])
	}
	return 0,true
}

func main(){
	var IP string
	var port uint64
	fmt.Print("请输入登录的ip地址和端口号（格式：127.0.0.1 300）：")
	fmt.Scanf("%s%d",&IP,&port)
	EntryNode,_:=newManageNode(IP,port)

	//开始监听，判断是否有节点请求加入集群或者是否有领导者发送的心跳
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

	//since:=time.Now()
	//electionTimeout:=EntryNode.ElectionTimeout
	//heartbeattimeout:=afterBetween(EntryNode.ElectionTimeout(),EntryNode.ElectionTimeout()*2)
	timer:=time.NewTicker(EntryNode.ElectionTimeout())

	for{
		update:=false
		data := make([]byte, 1024)
		n,_,err:=conn.ReadFromUDP(data)
		if err != nil {
			break
		}

		c:=make(chan []byte,1024)
		if data!=nil{
			c<-data
		}
		select {
		case <-c:
			//收到的消息类型：
			//1、第一个加入节点的网络
			//2、新节点加入集群的请求
			//3、旧节点下线重新加入集群
			//4、领导者的心跳
			msg_str := strings.Split(string(data[0:n]), "|")
			//处理监听到的数据
			switch msg_str[1] {
			case "joinCluster":
				//var joinresponse uint64
				joinresponse,join :=EntryNode.JoinCluster(msg_str)
				conn.Write([]byte(EntryNode.IP+"|"+"JoinCluster|"+strconv.Itoa(time.Now().Nanosecond())+"|"+strconv.Itoa(int(joinresponse))+"|"+strconv.FormatBool(join)+"|"+msg_str[n-1]))
				go EntryNode.nodeJoin(msg_str[n-1],joinresponse,join)
			case "AppendEntriesRequest":
				update=true
			}

		case <-timer.C:
			//超时未收到来自领导者的心跳，则认为未收到集群中领导者下线
			EntryNode.leader=" "
			update=true
		}

		if update{
			timer=time.NewTicker(EntryNode.ElectionTimeout())
		}


	}
}


func (s ManageNode)nodeJoin(nodeip string,nodeType uint64,nodeFlag bool){
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

	if nodeType==2&&nodeFlag==true{//如果是新节点加入集群则广播给集群中的所有节点且更新Nodes
		line:="JoinCluster|"+strconv.Itoa(time.Now().Nanosecond())+"|"+nodeip
		_, err = conn.Write([]byte(s.IP+line))
		s.Nodes=append(s.Nodes,nodeip)
	}
	if nodeType==1{//如果是第一个加入集群的节点，则还需要更新领导者
		s.Nodes=append(s.Nodes,nodeip)
		s.leader=nodeip
	}
}
