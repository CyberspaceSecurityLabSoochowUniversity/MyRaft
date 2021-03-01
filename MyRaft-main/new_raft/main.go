package new_raft

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

func main(){
	//打开监听端口
	conn, err := net.Dial("udp", "192.168.3.130：3000")  //打开监听端口
	if err != nil {
		fmt.Println("conn fail...")
	}
	defer conn.Close()

	//预先准备消息缓冲区
	buffer:=make([]byte,1024)
	//已知入口节点ip地址：192.168.3.130:3000
	//输入自己的ip地址
	server:=Server{}
	fmt.Printf("请输入服务器的IP和port(127.0.0.1:3000)：")
	fmt.Scanf("%s%d",&server.IP,&server.port)
	fmt.Printf("请输入服务器的名称：")
	fmt.Scanf("%s",&server.name)
	server.path= "c:\\desttop\\go"
	server.currentTerm=0
	server.state=Follower

	for{
		//发送加入集群请求
		conn.Write([]byte(server.IP+"|"+"JoinCluster|"+strconv.Itoa(time.Now().Nanosecond())+"|192.168.3.130"))
		n, err := conn.Read(buffer)//读取入口节点的回复
		if n == 0 || err != nil {
			continue
		}

		//解析协议
		msg_str := strings.Split(string(buffer[0:n]), "|")  //将从入口节点收到的字节流分段保存到msg_str这个数组中
		//msg_str[0]:加入节点的ip地址
		//msg_str[1]:JoinCluster
		//msg_str[2]:当前时间
		//msg_str[3]:节点类型（1，2新/3旧）
		//msg_str[4]:是否允许加入集群（true允许，false拒绝）
		//msg_str[5]:入口节点的ip
		switch msg_str[3]{
			case "1"://新节点加入且是集群中第一个节点
				server.setState(Leader)
				server.Leaderloop()
			case "3"://如果是旧节点应该相应的返回一些状态消息
				server.Followerloop()
			case "2":
				if msg_str[4]=="true"{
					server.Followerloop()
				}else{//如果拒绝加入集群，等待一段时间(例如选举时间)再向入口节点重新申请
					time.Sleep(200*time.Millisecond)
					break
				}
		}
	}
}