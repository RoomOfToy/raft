package raft

import "go.uber.org/zap"

type State interface {
	HandleAppendEntries(machine *Machine, msg Message)
	HandleAppendEntriesResponse(machine *Machine, msg Message)
	HandleRequestVote(machine *Machine, msg Message)
	HandleRequestVoteResponse(machine *Machine, msg Message)

	HandleElectionTimeout(machine *Machine)
	HandleLeaderTimeout(machine *Machine)
}

type LogEntry struct {
	term  int
	entry interface{}
}

func NewLogEntry(term int, entry interface{}) LogEntry {
	return LogEntry{
		term:  term,
		entry: entry,
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
}

func NewMachine() *Machine {
	return &Machine{
		control:      nil,
		term:         0,
		state:        Follower{},
		votedFor:     -1,
		votesGranted: 0,
		log:          []LogEntry{},
		commitIdx:    -1,
		lastApplied:  -1,
	}
}

func (m Machine) SetController(control *Controller) {
	m.control = control
}

func (m *Machine) AppendEntries(prevIdx, prevTerm int, entries []LogEntry) bool {
	if prevIdx+1 > len(m.log) {
		return false
	}

	if prevIdx >= 0 && m.log[prevIdx].term != prevTerm {
		return false
	}
	m.log = append(m.log[prevIdx+1:], entries...)
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
		Logger.Info("Handle message", zap.String("type", "AppendEntries"))
		m.state.HandleAppendEntries(m, msg)
	case AppendEntriesResponse:
		Logger.Info("Handle message", zap.String("type", "AppendEntriesResponse"))
		m.state.HandleAppendEntriesResponse(m, msg)
	case RequestVote:
		Logger.Info("Handle message", zap.String("type", "RequestVote"))
		m.state.HandleRequestVote(m, msg)
	case RequestVoteResponse:
		Logger.Info("Handle message", zap.String("type", "HandleRequestVoteResponse"))
		m.state.HandleRequestVoteResponse(m, msg)
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
		prevLogTerm = m.log[prevLogIdx].term
	}
	m.control.SendMessage(NewAppendEntries(m.control.addr, dest, m.term, prevLogIdx, prevLogTerm, m.log[prevLogIdx+1:], m.commitIdx))
}

func HandleRequestVote(machine *Machine, msg Message) {
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
	Logger.Info("HandleElectionTimeout", zap.Int("server", machine.control.addr))
	machine.state = Candidate{}
	machine.term += 1
	machine.votedFor = machine.control.addr // vote for itself
	machine.control.ResetElectionTimer()
	machine.votesGranted = 1

	for _, dest := range machine.control.peers {
		lastLogIdx, lastLogTerm := len(machine.log)-1, -1
		if lastLogIdx != -1 {
			lastLogTerm = machine.log[lastLogIdx].term
		}
		machine.control.SendMessage(NewRequestVote(machine.control.addr, dest, machine.term, lastLogIdx, lastLogTerm))
	}
}

type Follower struct{}

func (f Follower) HandleAppendEntries(machine *Machine, msg Message) {
	machine.votedFor = -1
	message := msg.(AppendEntries)
	logOk := message.prevLogIdx == -1 ||
		(message.prevLogIdx >= 0 && message.prevLogIdx < len(machine.log) &&
			message.prevLogTerm == machine.log[message.prevLogIdx].term)
	if message.term < machine.term || !logOk {
		// failure
		machine.control.SendMessage(NewAppendEntriesResponse(machine.control.addr, message.source, machine.term, false, -1))
	} else {
		// appending should work
		ok := machine.AppendEntries(message.prevLogIdx, message.prevLogTerm, message.entries)
		if !ok {
			Logger.Panic("Follower HandleAppendEntries Failure")
		}
		machine.control.SendMessage(NewAppendEntriesResponse(machine.control.addr, message.source, machine.term, true, message.prevLogIdx+len(message.entries)))
		machine.commitIdx = message.leaderCommit
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
	if message.term == machine.term {
		// convert to follower and handle message
		machine.state = Follower{}
		machine.HandleMessage(msg)
	}
}

func (c Candidate) HandleAppendEntriesResponse(machine *Machine, msg Message) {}

func (c Candidate) HandleRequestVote(machine *Machine, msg Message) {
	HandleRequestVote(machine, msg)
}

func (c Candidate) HandleRequestVoteResponse(machine *Machine, msg Message) {
	message := msg.(RequestVoteResponse)
	if message.term < machine.term {
		// ignore out of date message
	}

	if message.voteGranted == 1 {
		machine.votesGranted += 1
		if machine.votesGranted > machine.control.nservers / 2 {
			Logger.Warn("Machine become Leader", zap.Int("Addr", machine.control.addr))
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
	if message.success {
		machine.matchIdx[message.source] = message.matchIdx
		machine.nextIdx[message.source] = message.matchIdx+1

		// check for consensus on log entries
		matches := sortMapByValue(machine.matchIdx)
		machine.commitIdx = matches[len(machine.matchIdx) / 2].value
		if machine.lastApplied < machine.commitIdx {
			machine.control.ApplyEntries(machine.log[machine.lastApplied+1:machine.commitIdx+1])
			machine.lastApplied = machine.commitIdx
		}
	} else {
		// it failed for this follower
		// immediately retry with a lower nextIdx value
		machine.nextIdx[message.source] -= 1
		machine.SendAppendEntry(message.source)
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

