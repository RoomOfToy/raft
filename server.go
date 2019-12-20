package raft

func serve(addr int) {
	dispatcher := NewTransportDispatcher(addr)
	dispatcher.Start()
	m := NewMachine()
	controller := NewController(addr, dispatcher, m, nil)
	m.SetController(controller)
	controller.Start()
}
