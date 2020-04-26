package main

import (
	"sync"
)

// Replica struct
type Replica struct {
	IP          string 
	Database    map[string]string 
	Address     string
	Cell        []string
	Slots		[]Slot
	ToApply     Slot
	Listeners   map[string]chan string
	Lock        sync.Mutex
	Kill        chan struct{}
}

// Pair is a Key-Value pair
type Pair struct {
	Key   string
	Value string
}

// Slot is used when accepted
type Slot struct {
	//ID		int `json:"id"`
	Decided  	bool 		//`json:"decided"`
	Number		int			//`json:"command"`
	Sequence  	Sequence 	//`json:"sequence"`
	Command 	Command 	//`json:"accepted"`
}

// Command struct
type Command struct {
	Address   	string
	Command   	string
	Sequence	Sequence
	Key			string
	Tag       	int64
}

// Sequence struct
type Sequence struct {
	Number  int
	Address string
}

// Nothing struct
type Nothing struct{}

// PrepareRequest struct
type PrepareRequest struct {
	Slot     Slot
	Sequence Sequence
}

// Response struct
type Response struct {
	IsOkay  bool
	Promise Sequence
	Command Command
}

// AcceptResponse struct
type AcceptResponse struct {
}

// AcceptRequest struct
type AcceptRequest struct {
	Seq     Sequence
	Command Command
	Slot    Slot
}
// DecideRequest struct
type DecideRequest struct {
	Slot  Slot
	Value Command
}

// ProposeRequest struct
type ProposeRequest struct {
	Command Command
}

// Accept struct
type Accept struct {
	Slot    Slot
	Seq     Sequence
	Command Command
}

// AllRequests struct
type AllRequests struct {
	Prepare  PrepareRequest
	Accepted AcceptRequest
	Propose  ProposeRequest
	Accept   Accept
	Decide   DecideRequest
	Address  string
}
