package main

import (
	"sync"
)

// Replica struct
type Replica struct {
	IP          string 
	Port        string
	Database    map[string]string 
	// Slots       []Slot
	Address     string
	Cell        []string
	ToApply     int
	Listeners   map[string]chan string
	Lock        sync.Mutex
	Kill        chan struct{}
	HighestSlot Slot
}

// Pair is a Key-Value pair
type Pair struct {
	Key   string
	Value string
}

// Slot is used when accepted
type Slot struct {
	//ID		int `json:"id"`
	Decided  	bool 		`json:"decided"`
	Command  	Command 	`json:"command"`
	Promise  	Sequence 	`json:"sequence"`
	Accepted 	Command 	`json:"accepted"`
	HighestN 	int 		`json:"highest_n"`
}

// Command struct
type Command struct {
	SequenceN int
	Address   string
	Command   string
	Tag       int
}

// Sequence struct
type Sequence struct {
	Number  int
	Address string
}

// Nothing struct
type Nothing struct{}
