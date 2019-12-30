package raft

import (
	"fmt"
	"go.uber.org/zap"
)

type State interface {
	HandleAppendEntries(machine *Machine, msg Message)
	HandleAppendEntriesResponse(machine *Machine, msg Message)
	HandleRequestVote(machine *Machine, msg Message)
	HandleRequestVoteResponse(machine *Machine, msg Message)

	HandleElectionTimeout(machine *Machine)
	HandleLeaderTimeout(machine *Machine)
}

type LogEntry struct {
	Tm    int
	Entry interface{}
}

func (le LogEntry) String() string {
	return fmt.Sprintf("LogEntry{ Tm: %d, Entry: %+v }", le.Tm, le.Entry)
}

func NewLogEntry(term int, entry interface{}) LogEntry {
	return LogEntry{
		Tm:    term,
		Entry: entry,
	}
}

type Machine struct {
	control      *Controller
	term         int
	state        State
	votedFor     int
	votesGranted int

	log []LogEntry

	commitIdx   int
	lastApplied int

	nextIdx  map[int]int
	matchIdx map[int]int

	logger *zap.Logger
}

func NewMachine(logger *zap.Logger) *Machine {
	return &Machine{
		control:      nil,
		term:         0,
		state:        Follower{},
		votedFor:     -1,
		votesGranted: 0,
		log:          []LogEntry{},
		commitIdx:    -1,
		lastApplied:  -1,

		logger: logger,
	}
}

func (m Machine) SetController(control *Controller) {
	m.control = control
}

func (m *Machine) AppendEntries(prevIdx, prevTerm int, entries []LogEntry) bool {
	if prevIdx+1 > len(m.log) {
		return false
	}

	if prevIdx >= 0 && m.log[prevIdx].Tm != prevTerm {
		return false
	}
	m.log = append(m.log[:prevIdx+1], entries...)
	return true
}

func (m *Machine) RestLeader() {
	m.nextIdx = make(map[int]int, len(m.control.peers))
	m.matchIdx = make(map[int]int, len(m.control.peers))
	for _, p := range m.control.peers {
		m.nextIdx[p] = len(m.log)
		m.matchIdx[p] = -1
	}

	m.votedFor = -1
}

func (m *Machine) HandleMessage(msg Message) {
	if msg.Term() > m.term {
		m.term = msg.Term()
		m.state = Follower{}
		m.votedFor = -1
	}
	switch msg.(type) {
	case AppendEntries:
		m.logger.Info("Handle message", zap.String("type", "AppendEntries"))
		m.state.HandleAppendEntries(m, msg)
	case AppendEntriesResponse:
		if msg.Term() == m.term {
			m.logger.Info("Handle message", zap.String("type", "AppendEntriesResponse"))
			m.state.HandleAppendEntriesResponse(m, msg)
		}
	case RequestVote:
		m.logger.Info("Handle message", zap.String("type", "RequestVote"))
		m.state.HandleRequestVote(m, msg)
	case RequestVoteResponse:
		if msg.Term() == m.term {
			m.logger.Info("Handle message", zap.String("type", "HandleRequestVoteResponse"))
			m.state.HandleRequestVoteResponse(m, msg)
		}
	}
}

func (m *Machine) HandleElectionTimout() {
	m.state.HandleElectionTimeout(m)
}

func (m *Machine) HandleLeaderTimeout() {
	m.state.HandleLeaderTimeout(m)
}

func (m *Machine) AppendNewEntry(entry interface{}) {
	e := NewLogEntry(m.term, entry)
	m.log = append(m.log, e)
	m.logger.Info("AppendNewEntry", zap.String("entry", e.String()))
	m.SendAppendEntries()
	m.control.ResetLeaderTimeout()
}

func (m *Machine) SendAppendEntries() {
	for dest := range m.control.peers {
		m.SendAppendEntry(dest)
	}
}

func (m *Machine) SendAppendEntry(dest int) {
	prevLogIdx, prevLogTerm := m.nextIdx[dest]-1, -1
	if prevLogIdx >= 0 {
		prevLogTerm = m.log[prevLogIdx].Tm
	}
	m.control.SendMessage(NewAppendEntries(m.control.addr, dest, m.term, prevLogIdx, prevLogTerm, m.log[prevLogIdx+1:], m.commitIdx))
}

func HandleRequestVote(machine *Machine, msg Message) {
	machine.logger.Info("RequestVote", zap.Int("from", machine.control.addr), zap.Int("to", msg.Source()))
	if msg.Term() < machine.term ||
		(msg.Term() == machine.term && machine.votedFor != -1 && machine.votedFor != msg.Source()) {
		machine.control.SendMessage(NewRequestVoteResponse(machine.control.addr, msg.Source(), machine.term, 0))
	} else {
		machine.votedFor = msg.Source()
		machine.control.SendMessage(NewRequestVoteResponse(machine.control.addr, msg.Source(), machine.term, 1))
		machine.control.ResetElectionTimer()
	}
}

func HandleElectionTimeout(machine *Machine) {
	machine.logger.Info("HandleElectionTimeout", zap.Int("server", machine.control.addr))
	machine.state = Candidate{}
	machine.term += 1
	machine.votedFor = machine.control.addr // vote for itself
	machine.control.ResetElectionTimer()
	machine.votesGranted = 1

	for _, dest := range machine.control.peers {
		lastLogIdx, lastLogTerm := len(machine.log)-1, -1
		if lastLogIdx != -1 {
			lastLogTerm = machine.log[lastLogIdx].Tm
		}
		machine.control.SendMessage(NewRequestVote(machine.control.addr, dest, machine.term, lastLogIdx, lastLogTerm))
	}
}

type Follower struct{}

func (f Follower) HandleAppendEntries(machine *Machine, msg Message) {
	machine.votedFor = -1
	message := msg.(AppendEntries)
	logOk := message.PrevLogIdx == -1 ||
		(message.PrevLogIdx >= 0 && message.PrevLogIdx < len(machine.log) &&
			message.PrevLogTm == machine.log[message.PrevLogIdx].Tm)
	if message.Tm < machine.term || !logOk {
		// failure
		machine.control.SendMessage(NewAppendEntriesResponse(machine.control.addr, message.Src, machine.term, false, -1))
	} else {
		// appending should work
		ok := machine.AppendEntries(message.PrevLogIdx, message.PrevLogTm, message.Entries)
		if !ok {
			machine.logger.Panic("Follower HandleAppendEntries Failure")
			return
		}
		machine.control.SendMessage(NewAppendEntriesResponse(machine.control.addr, message.Src, machine.term, true, message.PrevLogIdx+len(message.Entries)))
		machine.commitIdx = message.LeaderCommit
		if machine.lastApplied < machine.commitIdx {
			machine.control.ApplyEntries(machine.log[machine.lastApplied+1 : machine.commitIdx+1])
			machine.lastApplied = machine.commitIdx
		}
		machine.control.ResetElectionTimer()
	}
}

func (f Follower) HandleAppendEntriesResponse(machine *Machine, msg Message) {}

func (f Follower) HandleRequestVote(machine *Machine, msg Message) {
	HandleRequestVote(machine, msg)
}

func (f Follower) HandleRequestVoteResponse(machine *Machine, msg Message) {}

func (f Follower) HandleElectionTimeout(machine *Machine) {
	HandleElectionTimeout(machine)
}

func (f Follower) HandleLeaderTimeout(machine *Machine) {}

type Candidate struct{}

func (c Candidate) HandleAppendEntries(machine *Machine, msg Message) {
	message := msg.(AppendEntries)
	if message.Tm == machine.term {
		// convert to follower and handle message
		machine.state = Follower{}
		machine.state.HandleAppendEntries(machine, msg)
	}
}

func (c Candidate) HandleAppendEntriesResponse(machine *Machine, msg Message) {}

func (c Candidate) HandleRequestVote(machine *Machine, msg Message) {
	HandleRequestVote(machine, msg)
}

func (c Candidate) HandleRequestVoteResponse(machine *Machine, msg Message) {
	machine.logger.Info("RequestVoteResponse", zap.Int("from", msg.Source()), zap.Int("to", machine.control.addr))
	message := msg.(RequestVoteResponse)
	if message.Tm < machine.term {
		// ignore out of date message
	}

	if message.VoteGranted == 1 {
		machine.votesGranted += 1
		if machine.votesGranted > machine.control.nservers/2 {
			machine.logger.Warn("Machine become Leader", zap.Int("Addr", machine.control.addr))
			machine.state = Leader{}
			machine.RestLeader()
			// upon leadership change, send and empty AppendEntries
			machine.SendAppendEntries()
			machine.control.ResetLeaderTimeout()
		}
	}
}

func (c Candidate) HandleElectionTimeout(machine *Machine) {
	HandleElectionTimeout(machine)
}

func (c Candidate) HandleLeaderTimeout(machine *Machine) {}

type Leader struct{}

func (l Leader) HandleAppendEntries(machine *Machine, msg Message) {}

func (l Leader) HandleAppendEntriesResponse(machine *Machine, msg Message) {
	// if the operation was successful, update leader settings for the follower
	message := msg.(AppendEntriesResponse)
	if message.Success {
		machine.matchIdx[message.Src] = message.MatchIdx
		machine.nextIdx[message.Src] = message.MatchIdx + 1

		// check for consensus on log entries
		matches := sortMapByValue(machine.matchIdx)
		machine.commitIdx = matches[len(machine.matchIdx)/2].value
		if machine.lastApplied < machine.commitIdx {
			machine.control.ApplyEntries(machine.log[machine.lastApplied+1 : machine.commitIdx+1])
			machine.lastApplied = machine.commitIdx
		}
	} else {
		// it failed for this follower
		// immediately retry with a lower nextIdx value
		machine.nextIdx[message.Src] -= 1
		machine.SendAppendEntry(message.Src)
	}
}

func (l Leader) HandleRequestVote(machine *Machine, msg Message) {
	HandleRequestVote(machine, msg)
}

func (l Leader) HandleRequestVoteResponse(machine *Machine, msg Message) {}

func (l Leader) HandleElectionTimeout(machine *Machine) {}

func (l Leader) HandleLeaderTimeout(machine *Machine) {
	// must send and append entries message to all followers
	machine.SendAppendEntries()
	// must reset the leader timeout
	machine.control.ResetLeaderTimeout()
}
