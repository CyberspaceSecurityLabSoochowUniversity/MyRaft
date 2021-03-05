package client

import (
	"testing"
)

func TestNewClient(t *testing.T)  {
	NewClient("127.0.0.1",10006,[]byte("10"))
}
