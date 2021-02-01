package main

import (
	raft "../raft-master-2"
	"fmt"
	"os"
	"time"
)

const(
	testElectionTimeout   = 200 * time.Millisecond
)


func main()  {
	transporter := raft.NewHTTPTransporter("/raft", testElectionTimeout)
	server,err := raft.NewServer("192.168.1.101","./tmp",transporter,nil,nil,"")
	if err != nil{
		fmt.Fprintln(os.Stdout,"create a new server err:",err.Error())
		return
	}
	fmt.Fprintf(os.Stdout,"%+v",server)
	server.Start()
	if _, err := server.Do(&raft.DefaultJoinCommand{Name: server.Name()}); err != nil {
		fmt.Fprintf(os.Stdout,"Server %s unable to join: %v", server.Name(), err)
	}
	defer server.Stop()
}
