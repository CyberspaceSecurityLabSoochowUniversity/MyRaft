package raft

import (
	"os"
	"path"
	"sync"
)

type Log struct {
	//ApplyFunc   func(*LogEntry, Command) (interface{}, error)
	file        *os.File
	path        string
	entries     []*LogEntry
	startIndex  uint64
	commitIndex uint64
	mutex       sync.RWMutex
	//startIndex  uint64
	//startTerm   uint64
	initialized bool
}

func NewLog(serverPath string) *Log{
	if serverPath == ""{
		return nil
	}
	l := &Log{
		path: path.Join(serverPath, "log"),
		entries: make([]*LogEntry, 0),
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
