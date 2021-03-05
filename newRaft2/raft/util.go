package raft

import (
	"math/rand"
	"time"
)

func afterBetween(min time.Duration, max time.Duration) <-chan time.Time {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	d, delta := min, max - min
	if delta > 0 {
		d += time.Duration(rand.Int63n(int64(delta)))
	}
	return time.After(d)
}

func afterBetween1(min time.Duration, max time.Duration) time.Duration {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	d, delta := min, max - min
	if delta > 0 {
		d += time.Duration(rand.Int63n(int64(delta)))
	}
	return d
}
