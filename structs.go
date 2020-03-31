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
	Decided  bool
	Command  Command
	Promise  Sequence
	Accepted Command
	HighestN int
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
