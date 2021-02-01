package raft

type LogEntry struct {
	//pb       *protobuf.LogEntry
	Position int64 // position in the log file
	log      *Log
	Index 	 uint
	Term     uint
	Key      string
	Value    string
}

func (l *Log) NewLogEntry(position int64,index uint,term uint,key string,value string) *LogEntry {
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
