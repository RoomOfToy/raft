package raft

import (
	"go.uber.org/zap"
	"io"
	"time"
)

type KVClient struct {
	logger *zap.Logger
	t      *Transport
}

func NewKVClient(logPath string) *KVClient {
	return &KVClient{
		logger: newLogger(LOGGER_LEVEL, logPath),
		t:      nil,
	}
}

func (kvc *KVClient) connect() {
	for {
		for k, v := range KV_SERVER_CONFIG {
			t := RunClient(v, kvc.logger)
			if t != nil {
				data, err := t.Read()
				if err != nil && err != io.EOF {
					kvc.logger.Error("Client: connect Read error", zap.Error(err))
					return
				}
				if err == io.EOF {
					return
				}
				c, err := decodeCmd(data)
				if err != nil {
					kvc.logger.Error("Client: connect decodeCmd error", zap.Error(err))
					return
				}
				if c.Name == "leader connected" {
					kvc.logger.Info("Connected to Leader KVServer", zap.Int("addr", k))
					kvc.t = t
					return
				} else {
					t.conn.Close()
					kvc.t = nil
				}
			}
		}
		kvc.logger.Warn("No KVServer available, retry after 2s")
		time.Sleep(2 * time.Second)
	}
}

func (kvc *KVClient) sendCmd(c CMD) *CMD {
	for {
		if kvc.t == nil {
			kvc.connect()
		}
		cBytes, err := encodeCmd(c)
		if err != nil {
			kvc.logger.Error("Client: sendCmd encodeCmd error", zap.Error(err))
			return nil
		}
		err = kvc.t.Send(cBytes)
		if err != nil {
			kvc.logger.Error("Client: sendCmd Send error", zap.Error(err))
			return nil
		}

		data, err := kvc.t.Read()
		if err != nil && err != io.EOF {
			kvc.logger.Error("Client: sendCmd Read error", zap.Error(err))
			return nil
		}
		if err == io.EOF {
			return nil
		}
		c, err := decodeCmd(data)
		if err != nil {
			kvc.logger.Error("Client: sendCmd decodeCmd error", zap.Error(err))
			return nil
		}
		return &c
	}
}

func (kvc *KVClient) decodeResp(c *CMD) (interface{}, bool) {
	if c.Name == "ok" {
		return c.Args, true
	} else {
		return c.Args, false
	}
}

func (kvc *KVClient) Get(key string) (interface{}, bool) {
	c := CMD{Name: "Get", Args: key}
	resp := kvc.sendCmd(c)
	return kvc.decodeResp(resp)
}

func (kvc *KVClient) Set(key string, value interface{}) (interface{}, bool) {
	c := CMD{Name: "Set", Args: Entry{Key: key, Value: value}}
	resp := kvc.sendCmd(c)
	return kvc.decodeResp(resp)
}
