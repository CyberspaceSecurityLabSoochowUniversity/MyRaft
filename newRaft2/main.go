package main

import (
	raft "./raft"
)


func main()  {
	//err, et := raft.NewEntrance("et1","192.168.1.101",raft.UdpPort,"first entrance")
	//if err != nil{
	//	fmt.Fprintln(os.Stderr,"NewEntrance Error")
	//	return
	//}
	//et.Start()

	s1,err := raft.NewServer("s2","../tmp","second server","192.168.1.102",10010,raft.UdpPort,nil)
	if err != nil{
		return
	}
	err = s1.Init("192.168.1.101",raft.UdpPort)
	if err != nil{
		return
	}
	err = s1.Start()
	if err != nil{
		return
	}
}
