package raft

import "time"

var (
	NSERVERS = 5

	LEADER_TIMEOUT = 1 * time.Second
	ELECTION_TIMEOUT = 3 * time.Second
	ELECTION_TIMEOUT_SPREAD = 500 * time.Millisecond

	RAFT_SERVER_CONFIG = map[int]string {
		0: "127.0.0.1:19000",
		1: "127.0.0.1:19001",
		2: "127.0.0.1:19002",
		3: "127.0.0.1:19003",
		4: "127.0.0.1:19004",
	}

	KV_SERVER_CONFIG = map[int]string {
		0: "127.0.0.1:20000",
		1: "127.0.0.1:20001",
		2: "127.0.0.1:20002",
		3: "127.0.0.1:20003",
		4: "127.0.0.1:20004",
	}
)
