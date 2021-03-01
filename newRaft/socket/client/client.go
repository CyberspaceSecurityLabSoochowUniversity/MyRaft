package client

import (
	"fmt"
	"net"
	"os"
	"strconv"
)

type Date struct {
	Id 		string
	Value   []byte
}

func NewClient(server_ip string,server_port int,message []byte) {
	serverAddr := server_ip + ":" + strconv.Itoa(server_port)
	addr,err := net.ResolveUDPAddr("udp",serverAddr)
	if err != nil{
		fmt.Fprintf(os.Stderr,"Client:New a upd client error:%s\n",err.Error())
		return
	}
	conn,err := net.DialUDP("udp",nil,addr)
	if err != nil{
		fmt.Fprintf(os.Stderr,"Client:Dial udp server error:%s\n",err.Error())
		return
	}
	defer conn.Close()

	_,err = conn.Write(message)
	if err != nil{
		fmt.Fprintf(os.Stderr,"Client:Write udp content error:%s\n",err.Error())
		return
	}
	//fmt.Fprintf(os.Stdout,"Client:Write:%s\n",str)
	//msg := make([]byte,server_recv_len)
	//n,err = conn.Read(msg)
	//if err != nil{
	//	fmt.Fprintf(os.Stderr,"Client:Read udp content error:%s\n",err.Error())
	//	return
	//}
	//fmt.Fprintf(os.Stdout,"Client:Response from server:%s\n",string(msg))


}

