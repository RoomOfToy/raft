package raft

import (
	"encoding/binary"
	"go.uber.org/zap"
	"io"
	"net"
)

// Transport with TLV protocol
//	12 bytes header (msg length) + msg body
type Transport struct {
	conn net.Conn
}

// NewTransport creates a Transport
func NewTransport(conn net.Conn) *Transport {
	return &Transport{conn}
}

// Send TLV msg
func (t *Transport) Send(msg []byte) error {
	// 12 bytes header: length of msg
	// rest: msg
	buf := make([]byte, 12+len(msg))
	binary.BigEndian.PutUint32(buf[:12], uint32(len(msg)))
	copy(buf[12:], msg)
	_, err := t.conn.Write(buf)
	if err != nil {
		return err
	}
	return nil
}

// Read TLV msg
func (t *Transport) Read() ([]byte, error) {
	// 12 bytes header indicates msg length
	header := make([]byte, 12)
	_, err := io.ReadFull(t.conn, header)
	if err != nil {
		return nil, err
	}
	dataLen := binary.BigEndian.Uint32(header)
	msg := make([]byte, dataLen)
	_, err = io.ReadFull(t.conn, msg)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func RunServer(addr string, handler func(t *Transport)) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		Logger.Error("Listen error", zap.String("addr", addr), zap.String("err", err.Error()))
		return
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			Logger.Error("Accept error", zap.String("err", err.Error()))
			continue
		}
		t := Transport{conn: conn}
		go handler(&t)
	}
}

func RunClient(serverAddr string) *Transport {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		Logger.Error("Connect error", zap.String("addr", serverAddr), zap.String("err", err.Error()))
		return nil
	}
	return &Transport{conn: conn}
}
