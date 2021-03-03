package main

import (
	raft "./raft"
	"fmt"
	"os"
)


func main()  {
	err, et := raft.NewEntrance("et1","192.168.1.100",raft.UdpPort,"first entrance")
	if err != nil{
		fmt.Fprintln(os.Stderr,"NewEntrance Error")
		return
	}
	go func() {
		et.Start()
	}()

	s1,err := raft.NewServer("s1","../tmp","first server","192.168.1.1",10010,raft.UdpPort,nil)
	if err != nil{
		return
	}
	err = s1.Init("192.168.1.100",raft.UdpPort)
	if err != nil{
		return
	}
	err = s1.Start()
	if err != nil{
		return
	}
}
