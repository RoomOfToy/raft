package raft

import (
	"bytes"
	"encoding/gob"
)

type Message interface {
	Source() int
	Dest() int
	Term() int
}

func init() {
	gob.Register(AppendEntries{})
	gob.Register(AppendEntriesResponse{})
	gob.Register(RequestVote{})
	gob.Register(RequestVoteResponse{})
}

func Encode(msg Message) ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(&msg); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decode(b []byte) (Message, error) {
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buf)
	var data Message
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}

type AppendEntries struct {
	Src          int
	Dst          int
	Tm           int
	PrevLogIdx   int
	PrevLogTm    int
	Entries      []LogEntry
	LeaderCommit int
}

func NewAppendEntries(source, dest, term, preLogIdx, preLogTerm int, entries []LogEntry, leaderCommit int) AppendEntries {
	return AppendEntries{
		Src:          source,
		Dst:          dest,
		Tm:           term,
		PrevLogIdx:   preLogIdx,
		PrevLogTm:    preLogTerm,
		Entries:      entries,
		LeaderCommit: leaderCommit,
	}
}

func (ap AppendEntries) Source() int {
	return ap.Src
}

func (ap AppendEntries) Dest() int {
	return ap.Dst
}

func (ap AppendEntries) Term() int {
	return ap.Tm
}

type AppendEntriesResponse struct {
	Src      int
	Dst      int
	Tm       int
	Success  bool
	MatchIdx int
}

func NewAppendEntriesResponse(source, dest, term int, success bool, matchIdx int) AppendEntriesResponse {
	return AppendEntriesResponse{
		Src:      source,
		Dst:      dest,
		Tm:       term,
		Success:  success,
		MatchIdx: matchIdx,
	}
}

func (ap AppendEntriesResponse) Source() int {
	return ap.Src
}

func (ap AppendEntriesResponse) Dest() int {
	return ap.Dst
}

func (ap AppendEntriesResponse) Term() int {
	return ap.Tm
}

type RequestVote struct {
	Src        int
	Dst        int
	Tm         int
	LastLogIdx int
	LastLogTm  int
}

func NewRequestVote(source, dest, term, lastLogIdx, lastLogTerm int) RequestVote {
	return RequestVote{
		Src:        source,
		Dst:        dest,
		Tm:         term,
		LastLogIdx: lastLogIdx,
		LastLogTm:  lastLogTerm,
	}
}

func (rv RequestVote) Source() int {
	return rv.Src
}

func (rv RequestVote) Dest() int {
	return rv.Dst
}

func (rv RequestVote) Term() int {
	return rv.Tm
}

type RequestVoteResponse struct {
	Src         int
	Dst         int
	Tm          int
	VoteGranted int
}

func NewRequestVoteResponse(source, dest, term, voteGranted int) RequestVoteResponse {
	return RequestVoteResponse{
		Src:         source,
		Dst:         dest,
		Tm:          term,
		VoteGranted: voteGranted,
	}
}

func (rvr RequestVoteResponse) Source() int {
	return rvr.Src
}

func (rvr RequestVoteResponse) Dest() int {
	return rvr.Dst
}

func (rvr RequestVoteResponse) Term() int {
	return rvr.Tm
}

var _ Message = (*AppendEntries)(nil)
var _ Message = (*AppendEntriesResponse)(nil)
var _ Message = (*RequestVote)(nil)
var _ Message = (*RequestVoteResponse)(nil)
