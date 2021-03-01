package new_raft

import (
"fmt"
"io"
"os"
"sync"
)

type Log struct{
	file  *os.File
	path  string
	entries []*LogEntry
	commitIndex  uint64
	mutex  sync.RWMutex
	startIndex  uint64
	startTerm   uint64
	initialized  bool
}

func newLog() *Log {
	return &Log{
		entries: make([]*LogEntry, 0),
	}
}

func (l *Log) CommitIndex() uint64 {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.commitIndex
}
func (l *Log) internalCurrentIndex() uint64 {		// 内部当前的索引
	if len(l.entries) == 0 {
		return l.startIndex
	}
	return l.entries[len(l.entries)-1].Index()
}

// The current index in the log.
func (l *Log) currentIndex() uint64 {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.internalCurrentIndex()
}

// The next index in the log.
func (l *Log) nextIndex() uint64 {
	return l.currentIndex() + 1
}

func (l *Log) isEmpty() bool {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return (len(l.entries) == 0) && (l.startIndex == 0)
}

// The name of the last command in the log.
func (l *Log) lastCommandName() string {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	if len(l.entries) > 0 {
		if entry := l.entries[len(l.entries)-1]; entry != nil {
			return entry.CommandName()
		}
	}
	return ""
}

// The current term in the log.
func (l *Log) currentTerm() uint64 {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if len(l.entries) == 0 {
		return l.startTerm
	}
	return l.entries[len(l.entries)-1].Term()
}

// Retrieves the last index and term that has been committed to the log.
func (l *Log) commitInfo() (index uint64, term uint64) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	// If we don't have any committed entries then just return zeros.
	if l.commitIndex == 0 {
		return 0, 0
	}

	// No new commit log after snapshot
	if l.commitIndex == l.startIndex {
		return l.startIndex, l.startTerm
	}

	// Return the last index & term from the last committed entry.
	entry := l.entries[l.commitIndex-1-l.startIndex]
	return entry.Index(), entry.Term()
}

// Retrieves the last index and term that has been appended to the log.
func (l *Log) lastInfo() (index uint64, term uint64) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	// If we don't have any entries then just return zeros.
	if len(l.entries) == 0 {
		return l.startIndex, l.startTerm
	}

	// Return the last index & term
	entry := l.entries[len(l.entries)-1]
	return entry.Index(), entry.Term()
}

// Updates the commit index
func (l *Log) updateCommitIndex(index uint64) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if index > l.commitIndex {
		l.commitIndex = index
	}
}

