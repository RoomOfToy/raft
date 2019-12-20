package raft

import (
	"bytes"
	"encoding/gob"
)

type Message interface {
	Source() int
	Dest() int
	Term() int

	Encode() ([]byte, error)
	Decode(b []byte) (Message, error)
}

type AppendEntries struct {
	source       int
	dest         int
	term         int
	prevLogIdx   int
	prevLogTerm  int
	entries      []LogEntry
	leaderCommit int
}

func NewAppendEntries(source, dest, term, preLogIdx, preLogTerm int, entries []LogEntry, leaderCommit int) AppendEntries {
	return AppendEntries{
		source:       source,
		dest:         dest,
		term:         term,
		prevLogIdx:   preLogIdx,
		prevLogTerm:  preLogTerm,
		entries:      entries,
		leaderCommit: leaderCommit,
	}
}

func (ap AppendEntries) Source() int {
	return ap.source
}

func (ap AppendEntries) Dest() int {
	return ap.dest
}

func (ap AppendEntries) Term() int {
	return ap.term
}

func (ap AppendEntries) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(ap); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (ap AppendEntries) Decode(b []byte) (Message, error) {
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buf)
	var data AppendEntries
	if err := decoder.Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

type AppendEntriesResponse struct {
	source   int
	dest     int
	term     int
	success  bool
	matchIdx int
}

func NewAppendEntriesResponse(source, dest, term int, success bool, matchIdx int) AppendEntriesResponse {
	return AppendEntriesResponse{
		source:   source,
		dest:     dest,
		term:     term,
		success:  success,
		matchIdx: matchIdx,
	}
}

func (ap AppendEntriesResponse) Source() int {
	return ap.source
}

func (ap AppendEntriesResponse) Dest() int {
	return ap.dest
}

func (ap AppendEntriesResponse) Term() int {
	return ap.term
}

func (ap AppendEntriesResponse) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(ap); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (ap AppendEntriesResponse) Decode(b []byte) (Message, error) {
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buf)
	var data AppendEntriesResponse
	if err := decoder.Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

type RequestVote struct {
	source      int
	dest        int
	term        int
	lastLogIdx  int
	lastLogTerm int
}

func NewRequestVote(source, dest, term, lastLogIdx, lastLogTerm int) RequestVote {
	return RequestVote{
		source:      source,
		dest:        dest,
		term:        term,
		lastLogIdx:  lastLogIdx,
		lastLogTerm: lastLogTerm,
	}
}

func (rv RequestVote) Source() int {
	return rv.source
}

func (rv RequestVote) Dest() int {
	return rv.dest
}

func (rv RequestVote) Term() int {
	return rv.term
}

func (rv RequestVote) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(rv); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (rv RequestVote) Decode(b []byte) (Message, error) {
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buf)
	var data RequestVote
	if err := decoder.Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}

type RequestVoteResponse struct {
	source      int
	dest        int
	term        int
	voteGranted int
}

func NewRequestVoteResponse(source, dest, term, voteGranted int) RequestVoteResponse {
	return RequestVoteResponse{
		source:      source,
		dest:        dest,
		term:        term,
		voteGranted: voteGranted,
	}
}

func (rvr RequestVoteResponse) Source() int {
	return rvr.source
}

func (rvr RequestVoteResponse) Dest() int {
	return rvr.dest
}

func (rvr RequestVoteResponse) Term() int {
	return rvr.term
}

func (rvr RequestVoteResponse) Encode() ([]byte, error) {
	var buf bytes.Buffer
	encoder := gob.NewEncoder(&buf)
	if err := encoder.Encode(rvr); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (rvr RequestVoteResponse) Decode(b []byte) (Message, error) {
	buf := bytes.NewBuffer(b)
	decoder := gob.NewDecoder(buf)
	var data RequestVoteResponse
	if err := decoder.Decode(data); err != nil {
		return nil, err
	}
	return data, nil
}
