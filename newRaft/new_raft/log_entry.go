package new_raft


import "protobuf"

type LogEntry struct { //LogEntry是日志项在内存中的描述结构，其最终存储在日志文件是经过protocol buffer编码以后的信息
	pb       *protobuf.LogEntry
	Position int64 // position in the log file  Position代表日志项存储在日志文件内的偏移
	log      *Log
	//event    *ev
}

func (e *LogEntry) Index() uint64 {
	return e.pb.GetIndex()
}

func (e *LogEntry) Term() uint64 {
	return e.pb.GetTerm()
}


