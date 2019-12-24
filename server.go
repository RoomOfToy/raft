package raft

func serve(addr int, logPath string) *Controller {
	logger := newLogger(LOGGER_LEVEL, logPath)
	dispatcher := NewTransportDispatcher(addr, logger)
	dispatcher.Start()
	m := NewMachine(logger)
	controller := NewController(addr, dispatcher, m, nil, logger)
	m.SetController(controller)
	return controller
}
