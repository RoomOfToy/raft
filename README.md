#

## Simple Raft Demo

check it by running test `TestServe`

it will print some log like the following:

```bash
{"level":"info","ts":1577090784.80836,"msg":"Starting server","addr":4}
{"level":"info","ts":1577090784.8083894,"msg":"Starting server","addr":0}
{"level":"info","ts":1577090784.8112876,"msg":"Starting server","addr":1}
{"level":"info","ts":1577090784.81105,"msg":"Starting server","addr":2}
{"level":"info","ts":1577090784.811084,"msg":"Starting server","addr":3}
{"level":"info","ts":1577090787.8235946,"msg":"HandleElectionTimeout","server":4}
{"level":"info","ts":1577090787.8399734,"msg":"Handle message","type":"RequestVote"}
{"level":"info","ts":1577090787.8400118,"msg":"RequestVote","from":1,"to":4}
{"level":"info","ts":1577090787.8559415,"msg":"Handle message","type":"RequestVote"}
{"level":"info","ts":1577090787.8559759,"msg":"RequestVote","from":2,"to":4}
{"level":"info","ts":1577090787.856052,"msg":"Handle message","type":"RequestVote"}
{"level":"info","ts":1577090787.85608,"msg":"RequestVote","from":3,"to":4}
{"level":"info","ts":1577090787.8568566,"msg":"Handle message","type":"HandleRequestVoteResponse"}
{"level":"info","ts":1577090787.8568692,"msg":"RequestVoteResponse","from":1,"to":4}
{"level":"info","ts":1577090787.8735092,"msg":"Handle message","type":"RequestVote"}
{"level":"info","ts":1577090787.8735385,"msg":"RequestVote","from":0,"to":4}
{"level":"info","ts":1577090787.8872962,"msg":"Handle message","type":"HandleRequestVoteResponse"}
{"level":"info","ts":1577090787.8873196,"msg":"RequestVoteResponse","from":3,"to":4}

{"level":"warn","ts":1577090787.8873239,"msg":"Machine become Leader","Addr":4}

{"level":"info","ts":1577090787.9030862,"msg":"Handle message","type":"HandleRequestVoteResponse"}
{"level":"info","ts":1577090787.903112,"msg":"Handle message","type":"HandleRequestVoteResponse"}
{"level":"info","ts":1577090787.9127061,"msg":"Handle message","type":"AppendEntries"}
{"level":"info","ts":1577090787.9156263,"msg":"Handle message","type":"AppendEntries"}
{"level":"info","ts":1577090787.9176967,"msg":"Handle message","type":"AppendEntries"}
{"level":"info","ts":1577090787.9148004,"msg":"Handle message","type":"AppendEntries"}
{"level":"info","ts":1577090787.9323075,"msg":"Handle message","type":"AppendEntriesResponse"}
{"level":"info","ts":1577090787.9443047,"msg":"Handle message","type":"AppendEntriesResponse"}
{"level":"info","ts":1577090787.9543014,"msg":"Handle message","type":"AppendEntriesResponse"}
{"level":"info","ts":1577090787.9871,"msg":"Handle message","type":"AppendEntriesResponse"}

...

{"level":"info","ts":1577090817.8091662,"msg":"HandleElectionTimeout","server":0}
{"level":"info","ts":1577090817.8308678,"msg":"Handle message","type":"RequestVote"}
{"level":"info","ts":1577090817.8309832,"msg":"RequestVote","from":2,"to":0}
{"level":"info","ts":1577090817.830813,"msg":"Handle message","type":"RequestVote"}
{"level":"info","ts":1577090817.8313496,"msg":"RequestVote","from":3,"to":0}
{"level":"info","ts":1577090817.8443682,"msg":"Handle message","type":"HandleRequestVoteResponse"}
{"level":"info","ts":1577090817.844395,"msg":"RequestVoteResponse","from":3,"to":0}
{"level":"info","ts":1577090817.8444004,"msg":"Handle message","type":"HandleRequestVoteResponse"}
{"level":"info","ts":1577090817.8444035,"msg":"RequestVoteResponse","from":2,"to":0}

{"level":"warn","ts":1577090817.8444066,"msg":"Machine become Leader","Addr":0}

{"level":"info","ts":1577090817.8592184,"msg":"Handle message","type":"AppendEntries"}
{"level":"info","ts":1577090817.8592682,"msg":"Handle message","type":"RequestVote"}
{"level":"info","ts":1577090817.8592732,"msg":"RequestVote","from":1,"to":0}
{"level":"info","ts":1577090817.8648024,"msg":"Handle message","type":"AppendEntries"}
{"level":"info","ts":1577090817.8668916,"msg":"Handle message","type":"AppendEntries"}
{"level":"info","ts":1577090817.87744,"msg":"Handle message","type":"HandleRequestVoteResponse"}
{"level":"info","ts":1577090817.899436,"msg":"Handle message","type":"AppendEntriesResponse"}
{"level":"info","ts":1577090817.8994703,"msg":"Handle message","type":"AppendEntriesResponse"}
{"level":"info","ts":1577090817.8997052,"msg":"Handle message","type":"AppendEntriesResponse"}

...
```

reference: [David Beazley repo raft_jun_2019](https://github.com/dabeaz/raft_jun_2019/)

i just port it from Python into Go, thanks to David, he is really a good teacher!!!
