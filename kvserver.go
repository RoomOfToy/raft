package raft

import (
	"bytes"
	"encoding/gob"
	"go.uber.org/zap"
	"io"
	"sync"
)

var KV_SERVER_CONFIG = map[int]string{
	0: "127.0.0.1:20100",
	1: "127.0.0.1:20101",
	2: "127.0.0.1:20102",
	3: "127.0.0.1:20103",
	4: "127.0.0.1:20104",
}

type kvStore struct {
	lock sync.RWMutex
	data map[string]interface{}
}

func newKVStore() *kvStore {
	return &kvStore{
		lock: sync.RWMutex{},
		data: make(map[string]interface{}),
	}
}

func (kv *kvStore) Get(key string) interface{} {
	kv.lock.RLock()
	defer kv.lock.RUnlock()
	return kv.data[key]
}

func (kv *kvStore) Set(key string, value interface{}) {
	kv.lock.Lock()
	defer kv.lock.Unlock()
	kv.data[key] = value
}

type KVServer struct {
	logger  *zap.Logger
	control *Controller
	store   *kvStore

	commitFlag chan struct{}
}

func NewKVServer(logger *zap.Logger, controller *Controller) *KVServer {
	kvs := &KVServer{
		logger:  logger,
		control: controller,
		store:   newKVStore(),

		commitFlag: make(chan struct{}, 1),
	}
	kvs.control.applicator = kvs
	return kvs
}

type Entry struct {
	Key   string
	Value interface{}
}

func init() {
	gob.Register(Entry{})
}

func (kvs *KVServer) ApplyEntries(entries []LogEntry) {
	kvs.logger.Info("KVStore: applying entries", zap.Int("addr", kvs.control.addr))
	for _, e := range entries {
		entry := e.Entry.(Entry)
		kvs.store.Set(entry.Key, entry.Value)
	}
	kvs.commitFlag <- struct{}{}
}

type CMD struct {
	Name string
	Args interface{}
}

func encodeCmd(c CMD) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(&c); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeCmd(b []byte) (CMD, error) {
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buf)
	var c CMD
	if err := decoder.Decode(&c); err != nil {
		return c, err
	}
	return c, nil
}

func (kvs *KVServer) handleCmd(c CMD) CMD {
	var res CMD
	switch c.Name {
	case "Get":
		res = CMD{
			Name: "ok",
			Args: kvs.store.Get(c.Args.(string)),
		}
	case "Set":
		entry := c.Args.(Entry)
		kvs.control.ApplyEntry(entry)
		// block here until commit
		<-kvs.commitFlag
		res = CMD{
			Name: "ok",
			Args: nil,
		}
	default:
		res = CMD{
			Name: "error",
			Args: "unknown command " + c.Name,
		}
	}
	return res
}

func (kvs *KVServer) HandleClient(t *Transport) {
	switch kvs.control.machine.state.(type) {
	case Leader:
		cBytes, err := encodeCmd(CMD{
			Name: "leader connected",
			Args: nil,
		})
		if err != nil {
			kvs.logger.Error("HandleClient: encodeCmd error", zap.Error(err))
		}
		err = t.Send(cBytes)
		if err != nil {
			kvs.logger.Error("HandleClient: transport Send error", zap.Error(err))
		}
		for {
			data, err := t.Read()
			if err != nil && err != io.EOF {
				kvs.logger.Error("HandleClient: transport Read error", zap.Error(err))
				return
			}
			if err == io.EOF {
				return
			}
			c, err := decodeCmd(data)
			if err != nil {
				kvs.logger.Error("HandleClient: decodeCmd error", zap.Error(err))
				return
			}
			handled := kvs.handleCmd(c)
			handledBytes, err := encodeCmd(handled)
			if err != nil {
				kvs.logger.Error("HandleClient: encodeCmd error after handleCmd", zap.Error(err))
				return
			}
			err = t.Send(handledBytes)
			if err != nil {
				kvs.logger.Error("HandleClient: transport Send error after handleCmd", zap.Error(err))
				return
			}
		}
	default:
		cBytes, err := encodeCmd(CMD{
			Name: "not leader error",
			Args: nil,
		})
		if err != nil {
			kvs.logger.Error("HandleClient: encodeCmd error", zap.Error(err))
		}
		err = t.Send(cBytes)
		if err != nil {
			kvs.logger.Error("HandleClient: transport Send error", zap.Error(err))
		}
		return
	}
}

func (kvs *KVServer) Start() {
	go RunServer(KV_SERVER_CONFIG[kvs.control.addr], kvs.logger, kvs.HandleClient)
	kvs.logger.Info("KVServer started", zap.Int("addr", kvs.control.addr))
}

func serveKV(addr int, logPath string, kvLogPath string) *KVServer {
	control := serve(addr, logPath)
	control.Start()
	logger := newLogger(LOGGER_LEVEL, kvLogPath)
	kv := NewKVServer(logger, control)
	kv.Start()
	return kv
}
