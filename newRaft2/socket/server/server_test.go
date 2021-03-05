package server

import (
	"testing"
)

func TestNewServer(t *testing.T)  {
	NewServer("127.0.0.1",10006,10)
}
