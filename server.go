package raft

func serve(addr int) *Controller {
	dispatcher := NewTransportDispatcher(addr)
	dispatcher.Start()
	m := NewMachine()
	controller := NewController(addr, dispatcher, m, nil)
	m.SetController(controller)
	return controller
}
