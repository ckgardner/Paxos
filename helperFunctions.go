package main

import (
	"fmt"
	"log"
	"net/rpc"
	"strings"
	"time"
)

// Propose function
func Propose(replica *Replica, item Slot) error {
	finished := false
	var nothing Nothing
	for !finished {
		fmt.Println("Starting Proposal Loop")
		fmt.Println("Highest slot:", replica.HighestSlot.HighestN)
		var data Slot
		data.HighestN = replica.HighestSlot.HighestN
		item.HighestN = replica.HighestSlot.HighestN
		replica.HighestSlot.Accepted.Command = item.Command.Command
		seen := -1
		completed := false
		for !completed {
			data.Command.Command = replica.HighestSlot.Accepted.Command
			item.HighestN = seen + 1
			data.HighestN = seen + 1
			replica.HighestSlot.Accepted.SequenceN = seen + 1
			totalOk := 0
			totalNot := 0
			for _, add := range replica.portAddress {
				time.Sleep(1000 * time.Millisecond) // latency
				var prepOk Slot
				fmt.Println("Asking", add)
				if err := call(add, "Replica.Prepare", data, &prepOk); err != nil {
					log.Printf("Bad prepare from %v", add)
				} else {
					if prepOk.Decided {
						totalOk++
					} else {
						totalNot++
						fmt.Printf("%v rejected \n", add)
						if prepOk.HighestN > seen {
							seen = prepOk.HighestN
						}
						if len(prepOk.Command.Command) > 0 {
							replica.HighestSlot.Accepted.Command = prepOk.Command.Command
						}
					}
				}
				if totalOk > len(replica.portAddress)/2 || totalNot > len(replica.portAddress)/2 {
					break
				}
			}
			if totalNot > len(replica.portAddress)/2 {
				fmt.Println("Majority declined... retry")
				time.Sleep(1000 * time.Millisecond) // latency
				continue
			}
			fmt.Println("Received a prepare majority. Accepting..")
			var prepOk Slot
			totalOk = 0
			totalNot = 0
			for _, add := range replica.portAddress {
				time.Sleep(1000 * time.Millisecond)
				if err := call(add, "Replica.Accept", data, &prepOk); err != nil {
					log.Printf("Bad decide call from %v", add)
				} else {
					if prepOk.Decided {
						totalOk++
					} else {
						fmt.Printf("%v rejected", add)
						totalNot++
					}
				}
			}
			if totalOk > len(replica.portAddress)/2 {
				fmt.Println("Got an accept majority. Deciding..")
				for _, add := range replica.portAddress {
					if err := call(add, "Replica.Decide", data, &nothing); err != nil {
						log.Printf("bad decide call from %v", add)
					}
				}
				if data.Command.Command == item.Command.Command {
					finished = true
				}
			} else {
				time.Sleep(1000 * time.Millisecond)
			}

			completed = true
		}
		replica.ToApply= -1
	}
	return nil
}

// Prepare is not an RPC
func (replica *Replica) Prepare(req Slot, res *Slot) error {

	replica.Lock.Lock()
	defer replica.Lock.Unlock()

	time.Sleep(1000 * time.Millisecond)
	log.Println("Prepare called with:", req.HighestN)

	if req.HighestN >= replica.ToApply  {
		if req.Command.SequenceN > replica.ToApply{
			fmt.Println(req.Command.SequenceN, replica.ToApply, " Was ACCPETED")
			replica.ToApply = req.HighestN
			req.Accepted.Command = req.Command.Command
			res.Decided = true
			
		} else{

			fmt.Println(req.Command.SequenceN, replica.ToApply, " Was REJECTED ")
			*res = replica.HighestSlot
			res.HighestN = replica.ToApply
			res.Command.Command = " "
			res.Decided = false
		}

	} else {

		res.Command.Command = replica.Cell[req.HighestN]
	
		if req.Command.Command == replica.Cell[req.HighestN]{

			res.Decided = true
			log.Println("ACCPETD:", req.HighestN, replica.Cell[req.HighestN])

		}else{

			res.Decided = false
			log.Println("REJECTED:", req.HighestN, replica.Cell[req.HighestN])
		}

	}

	time.Sleep(1000 * time.Millisecond)
	return nil
}

// Accept is not an RPC
func (replica *Replica) Accept(req Slot, res *Slot) error {
	replica.Lock.Lock()
	defer replica.Lock.Unlock()
	log.Println("=====", replica.Address, "Accepting..")
	
	if req.Command.SequenceN >= replica.HighestSlot.HighestN {

		log.Println("Sequence", req.Command.SequenceN , ">= highest promised", replica.HighestSlot.HighestN)
		replica.ToApply = req.Command.SequenceN 
		replica.HighestSlot.HighestN = req.Command.SequenceN
		replica.HighestSlot.Command.Command = req.Command.Command

		*res = replica.HighestSlot
		res.Decided = true
	} else {
		log.Println("Sequence", req.Command.SequenceN , "is NOT >= highest promised", replica.HighestSlot.HighestN)
		res.Decided = false
		*res = replica.HighestSlot
	}
	time.Sleep(1000 * time.Millisecond)
	return nil
}

// Decide is not an  RPC
func (replica *Replica) Decide(req Slot, res *Nothing) error {
	
	replica.Lock.Lock()
	defer replica.Lock.Unlock()
	time.Sleep(1000 * time.Millisecond)
	empty := "[EMPTY]"

	replica.ToApply = -1
	replica.HighestSlot.Command.SequenceN = -1

	commands := strings.Fields(req.Command.Command)
	if len(replica.Cell) == req.HighestN{
		
		replica.HighestSlot.HighestN ++
		if req.HighestN > 0 && replica.Cell[req.HighestN - 1] == empty{

			replica.Cell = append(replica.Cell,empty)
	
		}else{

			replica.Cell = append(replica.Cell,req.Command.Command)
	
		}
		switch commands[0]{
		
		case "put" :
			replica.Database[commands[1]] = commands[2]
			fmt.Printf("Adding to key/value pair: [%v] set to [%v]", commands[1], commands[2])
			
		case "get":
			value := replica.Database[commands[1]]
			fmt.Printf("Your key/values are: [%v] with [%v]", commands[1], value)
			
		case "delete":
			value := replica.Database[commands[1]]
			delete(replica.Database, commands[1])
			fmt.Printf("You deleted the key/value pair of : [%v] with [%v]", commands[1], value)
			
		}
	}else if len(replica.Cell) < req.HighestN{

		for i := 0; i < req.HighestN +1; i++{
			replica.Cell = append(replica.Cell, empty)
		}
	}else{

		if replica.Cell[req.HighestN] == empty{
			replica.Cell[req.HighestN] = req.Command.Command
			replica.HighestSlot.HighestN ++
			if replica.Cell[0] == empty{
				replica.HighestSlot.HighestN = 0
			}

		}
	}
				
	time.Sleep(1000 * time.Millisecond)
	return nil
}

func call(address string, method string, request interface{}, reply interface{}) error {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Printf("rpc.DialHTTP: %v", err)
		return err
	}

	defer client.Close()

	if err = client.Call(method, request, reply); err != nil {
		log.Printf("client.Call %s: %v", method, err)
		return err
	}

	return nil
}
