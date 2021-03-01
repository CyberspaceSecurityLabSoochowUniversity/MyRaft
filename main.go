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
	et.Start()



}
