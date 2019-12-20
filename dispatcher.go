package raft

import "go.uber.org/zap"

type Dispatcher interface {
	SendMsg(msg Message)
	RecvMsg(addr string) Message
}

type TransportDispatcher struct {
	addr       int
	nservers   int
	recvQueue  *DequeRW
	sendQueues []*DequeRW
}

func NewTransportDispatcher(addr int) *TransportDispatcher {
	nservers := len(RAFT_SERVER_CONFIG)
	sendQueues := make([]*DequeRW, nservers)
	for i := range sendQueues {
		sendQueues[i] = NewDequeRW(8)
	}
	return &TransportDispatcher{
		addr:       addr,
		nservers:   nservers,
		recvQueue:  NewDequeRW(8),
		sendQueues: sendQueues,
	}
}

func (td *TransportDispatcher) SendMsg(msg Message) {
	td.sendQueues[msg.Dest()].PushBack(msg)
}

func (td *TransportDispatcher) RecvMsg(addr int) Message {
	if addr != td.addr {
		Logger.Error("TransportDispatcher RecvMsg error: wrong addr", zap.Int("addr", addr))
		return nil
	}
	msg, err := td.recvQueue.PopFront()
	if err != nil {
		Logger.Debug("TransportDispatcher RecvMsg error: recvQueue error", zap.Error(err))
		return nil
	}
	return msg.(Message)
}

func (td *TransportDispatcher) RaftServer() {
	go RunServer(RAFT_SERVER_CONFIG[td.addr], td.raftReceiver)
}

func (td *TransportDispatcher) raftReceiver(t *Transport) {
	for {
		msg, err := t.Read()
		if err != nil {
			Logger.Error("TransportDispatcher RaftServer error: raftReceiver error", zap.Error(err))
			continue
		}
		td.recvQueue.PushBack(msg)
	}
}

func (td *TransportDispatcher) raftSender(addr int) {
	for {
		msg, err := td.sendQueues[addr].PopFront()
		if err != nil {
			Logger.Debug("TransportDispatcher raftSender error: Deque error", zap.Error(err))
			return
		}
		t := RunClient(RAFT_SERVER_CONFIG[addr])
		bytes, err := msg.(Message).Encode()
		if err != nil {
			Logger.Error("TransportDispatcher raftSender error: Encode error", zap.Error(err))
			return
		}
		err = t.Send(bytes)
		if err != nil {
			Logger.Error("TransportDispatcher raftSender error: Send error", zap.Error(err))
			return
		}
	}
}

func (td *TransportDispatcher) Start() {
	td.RaftServer()
	for i := 0; i < td.nservers; i++ {
		if i != td.addr {
			i := i
			go td.raftSender(i)
		}
	}
}
