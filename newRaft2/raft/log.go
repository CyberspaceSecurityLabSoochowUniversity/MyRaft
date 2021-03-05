package raft

import (
	client "../socket/client"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"time"
)

type Log struct {
	//ApplyFunc   func(*LogEntry, Command) (interface{}, error)
	file        	*os.File
	path        	string
	entries     	[]LogEntry
	startIndex  	uint64
	commitIndex 	uint64
	LastLogIndex    uint64
	LastLogTerm     uint64
	mutex      	 	sync.RWMutex
	//startIndex  uint64
	//startTerm   uint64
	initialized 	bool
}

func NewLog(serverPath string) *Log{
	if serverPath == ""{
		return nil
	}
	l := &Log{
		path: path.Join(serverPath, "log"),
		entries: make([]LogEntry, 0),
	}
	return l
}

func (l *Log)Path() string {
	return l.path
}

func (l *Log) isEmpty() bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return (len(l.entries) == 0) && (l.startIndex == 0)
}


type appendLogEntry struct {
	Name 			string
	ServerIp 		string
	ServerPort 		int
	LogIndex 		uint64
	Ip       		string
	Port     		int
}

func NewAppendLogEntry(name string,serverIp string,serverPort int,index uint64,
	ip string,port int) *appendLogEntry {
	ale := &appendLogEntry{
		Name: name,
		ServerIp: serverIp,
		ServerPort: serverPort,
		LogIndex: index,
		Ip: ip,
		Port: port,
	}
	return ale
}

func SendAppendLogEntryRequest(ale *appendLogEntry)  {
	if ale.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendAppendLogEntryRequest: IP is blank!")
		return
	}
	if ale.Port <= 0{
		fmt.Fprintln(os.Stdout,"SendAppendLogEntryRequest: Port is incorrect!")
		return
	}
	message,err := json.Marshal(ale)
	d := client.Date{Id: AppendLogEntryOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendAppendLogEntryRequest: Error converting data into Json!")
		return
	}

	go func() {
		for i:=0;i<3;i++{
			client.NewClient(ale.Ip,ale.Port,data)
			time.Sleep(20*time.Millisecond)
		}
	}()
}

func ReceiveAppendLogEntryRequest(message []byte) *appendLogEntry {
	ale := new(appendLogEntry)
	err := json.Unmarshal(message,&ale)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveAppendLogEntryRequest Error:",err.Error())
		return nil
	}
	return ale
}

type appendLogEntryResponse struct {
	Name string
	Entry []LogEntry
	Ip string
	Port int
}

func NewAppendLogEntryResponse(name string,entry []LogEntry,ip string,port int) *appendLogEntryResponse {
	alerp := &appendLogEntryResponse{
		Name: name,
		Entry: entry,
		Ip: ip,
		Port: port,
	}
	return alerp
}

func SendAppendLogEntryResponse(alerp *appendLogEntryResponse)  {
	if alerp.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendAppendLogEntryResponse: IP is blank!")
		return
	}
	if alerp.Port <= 0{
		fmt.Fprintln(os.Stdout,"SendAppendLogEntryResponse: Port is incorrect!")
		return
	}
	message,err := json.Marshal(alerp)
	d := client.Date{Id: AppendLogEntryResponseOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendAppendLogEntryResponse: Error converting data into Json!")
		return
	}
	//for i:=0;i<3;i++{
	//	client.NewClient(alerp.Ip,alerp.Port,data)
	//	time.Sleep(20*time.Millisecond)
	//}
	client.NewClient(alerp.Ip,alerp.Port,data)
}

func ReceiveAppendLogEntryResponse(message []byte) *appendLogEntryResponse {
	alerp := new(appendLogEntryResponse)
	err := json.Unmarshal(message,&alerp)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveAppendLogEntryResponse Error:",err.Error())
		return nil
	}
	return alerp
}
