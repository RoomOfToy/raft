package raft

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestKVServer(t *testing.T) {
	kvServers := make([]*KVServer, NSERVERS)
	for k := range RAFT_SERVER_CONFIG {
		kvServers[k] = serveKV(k, "./server-"+strconv.Itoa(k)+".log", "")
	}

	client := NewKVClient("")
	msg, ok := client.Set("a", "test")
	if !ok {
		t.Fatalf("%+v", msg)
	}
	msg, ok = client.Get("a")
	if !ok || msg.(string) != "test" {
		t.Fatalf("%+v", msg)
	}

	time.Sleep(ELECTION_TIMEOUT*2 + 1*time.Second)
}

func TestKVServer_RandomDown(t *testing.T) {
	kvServers := make([]*KVServer, NSERVERS)
	for k := range RAFT_SERVER_CONFIG {
		kvServers[k] = serveKV(k, "./server-"+strconv.Itoa(k)+".log", "")
	}
	ticker := time.NewTicker(15 * time.Second)
	sTicker := time.NewTicker(60 * time.Second)

	go func(t *testing.T) {
		kvs := []string{"a", "b", "c", "d", "e"}
		f := func(k string, v int, t *testing.T) func() {
			return func() {
				client := NewKVClient("")
				msg, ok := client.Set(k, v)
				if !ok {
					t.Fatalf("%+v", msg)
				}

				msg, ok = client.Get(k)
				if !ok || msg.(int) != v {
					t.Fatalf("%+v", msg)
				}
			}
		}
		for v, k := range kvs {
			f(k, v, t)()
			time.Sleep(10 * time.Second)
		}
	}(t)

Loop:
	for {
		select {
		case <-ticker.C:
			n := rand.Intn(NSERVERS)
			kvServers[n].control.pause()
			time.AfterFunc(5*time.Second, kvServers[n].control.resume)
		case <-sTicker.C:
			ticker.Stop()
			sTicker.Stop()
			break Loop
		}
	}
}
