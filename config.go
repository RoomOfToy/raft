package raft

import "time"

var (
	NSERVERS     = 5
	LOGGER_LEVEL = "info"

	LEADER_TIMEOUT          = 1 * time.Second
	ELECTION_TIMEOUT        = 3 * time.Second
	ELECTION_TIMEOUT_SPREAD = 100 * time.Millisecond

	RAFT_SERVER_CONFIG = map[int]string{
		0: "127.0.0.1:19100",
		1: "127.0.0.1:19101",
		2: "127.0.0.1:19102",
		3: "127.0.0.1:19103",
		4: "127.0.0.1:19104",
	}
)
