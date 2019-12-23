package raft

import (
	"go.uber.org/zap"
	"math/rand"
	"time"
)

type ControllerBase interface {
	ApplyEntries(entries []LogEntry)
}

type Controller struct {
	addr       int
	dispatcher *TransportDispatcher
	machine    *Machine
	applicator ControllerBase

	peers      []int
	nservers   int
	eventQueue *DequeRW
	running    bool
	paused     bool

	debugLog *zap.Logger

	leaderDeadline   time.Time
	electionDeadline time.Time
}

func NewController(addr int, dispatcher *TransportDispatcher, machine *Machine, applicator ControllerBase) *Controller {
	peers := make([]int, 0, dispatcher.nservers)
	for i := 0; i < dispatcher.nservers; i++ {
		if i != addr {
			peers = append(peers, i)
		}
	}
	c := &Controller{
		addr:       addr,
		dispatcher: dispatcher,
		machine:    machine,
		applicator: applicator,
		peers:      peers,
		nservers:   dispatcher.nservers,
		eventQueue: NewDequeRW(8),
		running:    false,
		paused:     false,
		debugLog:   Logger,
	}
	machine.control = c
	return c
}

func (c *Controller) SendMessage(msg Message) {
	c.debugLog.Info("SendMessage", zap.Int("from", c.addr), zap.Int("to", msg.Dest()))
	c.dispatcher.SendMsg(msg)
}

func (c *Controller) ApplyEntries(entries []LogEntry) {
	c.debugLog.Debug("ApplyEntries")
	if c.applicator != nil {
		c.applicator.ApplyEntries(entries)
	}
}

func (c *Controller) Start() {
	c.running = true
	c.debugLog.Info("Starting server", zap.Int("addr", c.addr))
	go c.run()
	go c.runReceiver()
	go c.runLeaderTimer()
	go c.runElectionTimer()
}

type event struct {
	name string
	args interface{}
}

func (c *Controller) run() {
	for c.running {
		evt, err := c.eventQueue.PopFront()
		if err != nil {
			c.debugLog.Debug("event error", zap.Error(err))
		}
		if !c.paused && evt != nil {
			e := evt.(event)
			switch e.name {
			case "HandleMessage":
				if e.args != nil {
					c.debugLog.Info("Event: HandleMessage", zap.Int("from", e.args.(Message).Source()), zap.Int("to", e.args.(Message).Dest()))
					c.machine.HandleMessage(e.args.(Message))
				}
			case "HandleLeaderTimeout":
				c.debugLog.Info("Event: HandleLeaderTimeout", zap.Int("server", c.addr))
				c.machine.HandleLeaderTimeout()
			case "HandleElectionTimeout":
				c.debugLog.Info("Event: HandleElectionTimout", zap.Int("server", c.addr))
				c.machine.HandleElectionTimout()
			case "AppendNewEntry":
				c.debugLog.Info("Event: AppendNewEntry", zap.Int("server", c.addr))
				c.machine.AppendNewEntry(e.args.(LogEntry))
			default:
				c.debugLog.Panic("Event: Unsupported Event", zap.String("name", e.name))
			}
		}
	}
}

func (c *Controller) runReceiver() {
	for c.running {
		msg := c.dispatcher.RecvMsg(c.addr)
		c.eventQueue.PushBack(event{"HandleMessage", msg})
	}
}

func (c *Controller) runLeaderTimer() {
	c.leaderDeadline = time.Now()
	for c.running {
		delay := c.leaderDeadline.Sub(time.Now())
		if delay <= 0 {
			delay = LEADER_TIMEOUT
			c.leaderDeadline = time.Now().Add(delay)
		}
		time.Sleep(delay)
		if time.Now().After(c.leaderDeadline) {
			c.eventQueue.PushBack(event{
				name: "HandleLeaderTimeout",
				args: nil,
			})
		}
	}
}

func (c *Controller) ResetLeaderTimeout() {
	c.debugLog.Info("ResetLeaderTimeout", zap.Int("addr", c.addr))
	c.leaderDeadline = time.Now().Add(LEADER_TIMEOUT)
}

func (c *Controller) runElectionTimer() {
	c.electionDeadline = time.Now()
	for c.running {
		delay := c.electionDeadline.Sub(time.Now())
		if delay <= 0 {
			c.newElectionDeadline()
		}
		time.Sleep(c.electionDeadline.Sub(time.Now()))
		if time.Now().After(c.electionDeadline) {
			c.eventQueue.PushBack(event{
				name: "HandleElectionTimeout",
				args: nil,
			})
		}
	}
}

func (c *Controller) newElectionDeadline() {
	newDeadline := time.Now().Add(time.Duration(rand.Int63n(10)) * ELECTION_TIMEOUT_SPREAD).Add(ELECTION_TIMEOUT)
	diff := newDeadline.Sub(c.electionDeadline)
	c.debugLog.Info("new election timeout", zap.Duration("diff", diff))  // seconds
	if diff > 0 {
		c.electionDeadline = newDeadline
	}
}

func (c *Controller) ResetElectionTimer() {
	c.debugLog.Info("ResetElectionTimer", zap.Int("addr", c.addr))
	c.newElectionDeadline()
}
