package raft

import (
	"go.uber.org/zap"
	"math/rand"
	"sync"
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

	leaderDeadline   *deadline
	electionDeadline *deadline
}

type deadline struct {
	lock sync.RWMutex
	t    time.Time
}

func newDeadline() *deadline {
	return &deadline{
		lock: sync.RWMutex{},
		t:    time.Now(),
	}
}

func (dl *deadline) get() time.Time {
	dl.lock.RLock()
	defer dl.lock.RUnlock()
	return dl.t
}

func (dl *deadline) set(t time.Time) {
	dl.lock.Lock()
	defer dl.lock.Unlock()
	dl.t = t
}

func NewController(addr int, dispatcher *TransportDispatcher, machine *Machine, applicator ControllerBase, logger *zap.Logger) *Controller {
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
		debugLog:   logger,

		leaderDeadline:   newDeadline(),
		electionDeadline: newDeadline(),
	}
	machine.control = c
	return c
}

func (c *Controller) SendMessage(msg Message) {
	// c.debugLog.Info("SendMessage", zap.Int("from", c.addr), zap.Int("to", msg.Dest()))
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
			// Logger.Warn("Event", zap.String("name", e.name))
			switch e.name {
			case "HandleMessage":
				// Logger.Warn("Event args", zap.Any("args", e.args))
				if e.args != nil {
					c.debugLog.Debug("Event: HandleMessage", zap.Int("from", e.args.(Message).Source()), zap.Int("to", e.args.(Message).Dest()))
					c.machine.HandleMessage(e.args.(Message))
				}
			case "HandleLeaderTimeout":
				c.debugLog.Debug("Event: HandleLeaderTimeout", zap.Int("server", c.addr))
				c.machine.HandleLeaderTimeout()
			case "HandleElectionTimeout":
				c.debugLog.Debug("Event: HandleElectionTimout", zap.Int("server", c.addr))
				c.machine.HandleElectionTimout()
			case "AppendNewEntry":
				c.debugLog.Debug("Event: AppendNewEntry", zap.Int("server", c.addr))
				c.machine.AppendNewEntry(e.args.(LogEntry))
			default:
				c.debugLog.Panic("Event: Unsupported Event", zap.String("name", e.name))
			}
		}
	}
}

func (c *Controller) pause() {
	c.paused = true
}

func (c *Controller) resume() {
	c.paused = false
}

func (c *Controller) runReceiver() {
	for c.running {
		msg := c.dispatcher.RecvMsg(c.addr)
		c.eventQueue.PushBack(event{"HandleMessage", msg})
	}
}

func (c *Controller) runLeaderTimer() {
	c.leaderDeadline.set(time.Now())
	for c.running {
		delay := c.leaderDeadline.get().Sub(time.Now())
		if delay <= 0 {
			delay = LEADER_TIMEOUT
			c.leaderDeadline.set(time.Now().Add(delay))
		}
		time.Sleep(delay)
		if time.Now().After(c.leaderDeadline.get()) {
			c.eventQueue.PushBack(event{
				name: "HandleLeaderTimeout",
				args: nil,
			})
		}
	}
}

func (c *Controller) ResetLeaderTimeout() {
	c.debugLog.Debug("ResetLeaderTimeout", zap.Int("addr", c.addr))
	c.leaderDeadline.set(time.Now().Add(LEADER_TIMEOUT))
}

func (c *Controller) runElectionTimer() {
	c.electionDeadline.set(time.Now())
	for c.running {
		delay := c.electionDeadline.get().Sub(time.Now())
		if delay <= 0 {
			c.newElectionDeadline()
		}
		time.Sleep(c.electionDeadline.get().Sub(time.Now()))
		if time.Now().After(c.electionDeadline.get()) {
			c.eventQueue.PushBack(event{
				name: "HandleElectionTimeout",
				args: nil,
			})
		}
	}
}

func (c *Controller) newElectionDeadline() {
	newDeadline := time.Now().Add(time.Duration(rand.Int63n(10)) * ELECTION_TIMEOUT_SPREAD).Add(ELECTION_TIMEOUT)
	diff := newDeadline.Sub(c.electionDeadline.get())
	c.debugLog.Debug("new election timeout", zap.Duration("diff", diff)) // seconds
	if diff > 0 {
		c.electionDeadline.set(newDeadline)
	}
}

func (c *Controller) ResetElectionTimer() {
	c.debugLog.Debug("ResetElectionTimer", zap.Int("addr", c.addr))
	c.newElectionDeadline()
}
