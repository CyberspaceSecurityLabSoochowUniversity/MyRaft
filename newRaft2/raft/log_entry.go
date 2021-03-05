package raft

import (
	client "../socket/client"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type LogEntry struct {
	//pb       *protobuf.LogEntry
	Position int64 // position in the log file
	log      *Log
	Index 	 uint64
	Term     uint64
	Key      string
	Value    string
}

func NewLogEntry(l *Log,position int64,index uint64,term uint64,key string,value string) *LogEntry {
	logEntry := &LogEntry{
		Position: position,
		Index: index,
		log: l,
		Term: term,
		Key: key,
		Value: value,
	}
	return logEntry
}

type AddLogEntry struct {
	Sign    uint64			//添加的日志的唯一标识
	Key 	string
	Value 	string
	Ip      string
	Port	int
}

func NewAddLogEntry(sign uint64,key string,value string,ip string,port int) *AddLogEntry {
	addle := &AddLogEntry{
		Sign: sign,
		Key: key,
		Value: value,
		Ip: ip,
		Port: port,
	}
	return addle
}

func SendAddLogEntryRequest(addle *AddLogEntry)  {
	if addle.Ip == ""{
		fmt.Fprintln(os.Stdout,"SendAddLogEntryRequest: IP is blank!")
		return
	}
	if addle.Port <= 0{
		fmt.Fprintln(os.Stdout,"SendAddLogEntryRequest: Port is incorrect!")
		return
	}
	if addle.Key == ""{
		fmt.Fprintln(os.Stdout,"SendAddLogEntryRequest Error: Key is blank!")
		return
	}
	message,err := json.Marshal(addle)
	d := client.Date{Id: AddLogEntryOrder,Value: message}
	data,err := json.Marshal(d)
	if err != nil{
		fmt.Fprintln(os.Stdout,"SendAddLogEntryRequest: Error converting data into Json!")
		return
	}
	for i:=0;i<3;i++{
		client.NewClient(addle.Ip,addle.Port,data)
		time.Sleep(22*time.Millisecond)
	}

}

func ReceiveAddLogEntryRequest(message []byte) *AddLogEntry {
	addle := new(AddLogEntry)
	err := json.Unmarshal(message,&addle)
	if err != nil{
		fmt.Fprintln(os.Stdout,"ReceiveAddLogEntryRequest Error:",err.Error())
		return nil
	}
	return addle
}
