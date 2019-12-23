package raft

import (
	"math/rand"
	"testing"
	"time"
)

func TestServe(t *testing.T) {
	controllers := make([]*Controller, NSERVERS)
	ticker := time.NewTicker(15 * time.Second)
	sTicker := time.NewTicker(60 * time.Second)
	for k, _ := range RAFT_SERVER_CONFIG {
		controllers[k] = serve(k)
		go controllers[k].Start()
	}
Loop:
	for {
		select {
		case <-ticker.C:
			n := rand.Intn(NSERVERS)
			controllers[n].pause()
			time.AfterFunc(5*time.Second, controllers[n].resume)
		case <-sTicker.C:
			ticker.Stop()
			sTicker.Stop()
		default:
			break Loop
		}
	}
}
