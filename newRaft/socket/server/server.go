package server

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

func NewServer(server_ip string,server_port int,server_recv_len int)  {
	address := server_ip + ":" + strconv.Itoa(server_port)
	addr,err := net.ResolveUDPAddr("udp",address)
	if err != nil{
		fmt.Fprintf(os.Stderr,"Server(%s):New a upd server error:%s\n",address,err.Error())
		return
	}
	conn,err := net.ListenUDP("udp",addr)
	if err != nil{
		fmt.Fprintf(os.Stderr,"Server(%s):New a listenupd error:%s\n",address,err.Error())
		return
	}
	defer conn.Close()
	for{
		//可以主动的读也可以主动的写
		data := make([]byte, server_recv_len)
		_,rAddr,err := conn.ReadFromUDP(data)
		if err != nil{
			fmt.Fprintf(os.Stderr,"Server(%s):Read udp content error:%s\n",address,err.Error())
			continue
		}

		//这里需要根据接收内容类型进行相应处理
		strData := string(data)
		fmt.Fprintf(os.Stdout,"Server(%s):Received:%s\n",address,strData)
		response := "receive success"
		_,err = conn.WriteToUDP([]byte(response),rAddr)
		if err != nil{
			fmt.Fprintf(os.Stderr,"Server(%s):response to client error:%s\n",address,err.Error())
			continue
		}
		fmt.Fprintf(os.Stdout, "Server(%s):Send success!\n", address)

		//需要一个命令可以退出该循环
	}
}
