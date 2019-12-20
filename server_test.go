package raft

import (
	"testing"
	"time"
)

func TestServe(t *testing.T) {
	for k, _ := range RAFT_SERVER_CONFIG {
		go serve(k)
	}
	time.Sleep(60 * time.Second)
}
