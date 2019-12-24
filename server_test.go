package raft

import (
	"bufio"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestServe(t *testing.T) {
	controllers := make([]*Controller, NSERVERS)
	ticker := time.NewTicker(15 * time.Second)
	sTicker := time.NewTicker(60 * time.Second)
	for k := range RAFT_SERVER_CONFIG {
		controllers[k] = serve(k, "./server-"+strconv.Itoa(k)+".log")
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
			break Loop
		}
	}
}

func TestNewLogger(t *testing.T) {
	logPath := "./test.log"
	logger := newLogger("info", logPath)
	logger.Debug("test0")
	logger.Info("test1")
	logger.Warn("test2")

	if _, err := os.Stat(logPath); err != nil {
		if os.IsNotExist(err) {
			t.Fatal("log file does NOT exist")
		}
	}

	log, err := os.Open(logPath)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		log.Close()
		os.Remove(logPath)
	}()

	reader := bufio.NewReader(log)
	var line string
	cnt := 0
	for {
		line, err = reader.ReadString('\n')
		if strings.Contains(line, "test0") {
			t.Fatal("log level NOT correct")
		}
		if cnt == 0 && !strings.Contains(line, "test1") {
			t.Fatal("log content NOT correct")
		}
		if cnt == 1 && !strings.Contains(line, "test2") {
			t.Fatal("log content NOT correct")
		}
		if err != nil {
			break
		}
		cnt++
	}
}
