package main
import(
	"log"
	"time"
	"strings"
	"fmt"
	"strconv"
	"net/rpc"
)
// Propose function
func(replica *Replica) Propose(req *DecideRequest, item Slot) error{
	finished := false
	for !finished{
		fmt.Println("Starting Proposal Loop")
		fmt.Println("Highest slot:", replica.Slots[req.Slot].Promise)
		var data Slot
		data.Decided = replica.Slots[req.Slot].Decided
		item.Decided = replica.Slots[req.Slot].Decided
		replica.Slots[req.Slot].Accepted.Command = item.Command.Command
		seen := -1
		completed := false
		for !completed{
			data.Command.Command = replica.Slots[req.Slot].Accepted.Command
			item.HighestN = seen + 1
			data.HighestN = seen + 1
			replica.Slots[req.Slot].Accepted.SequenceN = seen + 1
			totalOk := 0
			totalNot := 0
			for _, add := range replica.Cell{
				time.Sleep(1000 * time.Millisecond)
				var prepOk Slot
				fmt.Println("Asking", add)
				if err := call(add, "Node.Prepare", data, &prepOk); err != nil{
					log.Printf("Bad prepare from %v", add)
				} else{
					if prepOk.Decided{
						totalOk ++
					}else{
						totalNot ++
						fmt.Printf("%v rejected", add)
						if prepOk.HighestN > seen{
							seen = prepOk.HighestN
						}
						if len(prepOk.Command.Command) > 0{
							replica.Slots[req.Slot].Accepted.Command = prepOk.Command.Command
						}
					}
				}
				if totalOk > len(replica.Cell)/2 || totalNot > len(replica.Cell)/2{
					break
				}
			}
			if totalNot > len(replica.Cell)/2{
				fmt.Println("Majority declined... retry")
				time.Sleep(1000 * time.Millisecond)
				continue
			}
			fmt.Println("Received a prepare majority. Accepting..")
			var prepOk Slot
			totalOk = 0
			totalNot = 0
			for _, add := range replica.Cell{
				time.Sleep(1000 * time.Millisecond)
				if err := call(add, "Node.Accept", data, &prepOk); err != nil{
					log.Printf("Bad decide call from %v", add)
				}else{
					if prepOk.Decided {
						totalOk ++
					} else{
						fmt.Printf("%v rejected", add)
						totalNot ++
					}
				}
			}
			if totalOk > len(replica.Cell)/2{
				fmt.Println("Got an accept majority. Deciding..")
				for _, add := range replica.Cell{
					if err := call(add, "Node.Decide", data, struct{}{}); err != nil{
						log.Printf("bad decide call from %v", add)
					}
				}
				if data.Command.Command == item.Command.Command{
					completed = true
				}
			} else{
				time.Sleep(1000 * time.Millisecond)
			}
			completed = true
		}
		replica.Slots[req.Slot].HighestN = -1
	}
	return nil
}

// Prepare is not an RPC
func(replica *Replica) Prepare(req *PrepareRequest, res *PrepareResponse){
	replica.Lock.Lock()
	defer replica.Lock.Unlock()
	time.Sleep(1000 * time.Millisecond)
	log.Println("Prepare called with:", req.Slot, req.Seq)
	if len(replica.Slots) <= req.Slot{
		for i := len(replica.Slots); i <= req.Slot; i++{
			var newSlot Slot
			newSlot.Promise.Number = 0
			newSlot.HighestN = i
			replica.Slots = append(replica.Slots, newSlot)
		}
	}
	slot := replica.Slots[req.Slot]
	if slot.Decided{
		log.Println("This slot has already been decided", req.Slot)
	}
	res.Okay = req.Seq.Number > slot.Promise.Number
	if res.Okay{
		res.Promise = req.Seq
		replica.Slots[req.Slot].Promise = req.Seq
		res.Command = slot.Command
		if slot.Decided{
			log.Println("Prepare promising with command:", req.Slot, req.Seq, slot.Command)
		}else{
			log.Panicln("Prepare promising without a command:", req.Slot, req.Seq)
		}
	}else{
		log.Println("Preapre is rejecting because it has already promised", req.Slot, req.Seq, slot.Promise)
	}
	time.Sleep(1000 * time.Millisecond)
	return
}

// Accept is not an RPC
func(replica *Replica) Accept(req *PrepareRequest, res *PrepareResponse){
	replica.Lock.Lock()
	defer replica.Lock.Unlock()
	log.Println("=====", replica.Address, "Accepting..")
	slot := replica.Slots[req.Slot]
	if req.Seq.Number >= slot.Promise.Number{
		log.Println("Sequence", req.Seq.Number, ">= highest promised", slot.Promise.Number)
		slot.Promise.Number = req.Seq.Number
		slot.Accepted.SequenceN = req.Seq.Number
		slot.Command = res.Command

		res.Promise = slot.Promise
		res.Okay = true
	}else {
		log.Println("Sequence", req.Seq.Number, "is NOT >= highest promised", slot.Promise.Number)
		res.Okay = false
		res.Promise = slot.Promise
	}
	time.Sleep(1000 * time.Millisecond)
	return
}

// Decide is not an  RPC
func(replica *Replica) Decide(req *DecideRequest, res *Nothing){
	replica.Lock.Lock()
	defer replica.Lock.Unlock()
	time.Sleep(1000 * time.Millisecond)
	decisionmap := make(map[string]chan(string))

	log.Printf("[%v] Decide: called with Command- %v", req.Slot, req.Command)
	log.Printf("Requests promise - %v", req.Command.SequenceN)
	log.Printf("Slots promise - %v", replica.Slots[req.Slot].Promise)

	if replica.Slots[req.Slot].Decided && replica.Slots[req.Slot].Command.Command != req.Command.Command{
		log.Fatalf("[%v] Decide: --> PANIC, Decision was contradicted, quitting program", req.Slot)
	}
	replica.Slots[req.Slot].Promise.Number = req.Command.SequenceN
	if !replica.Slots[req.Slot].Decided {
		replica.Slots[req.Slot].Command = req.Command
		replica.Slots[req.Slot].Decided = true
		var t bool
		for i := req.Slot; i < len(replica.Slots); i++{
			if req.Slot > 0{
				t = replica.Slots[i-1].Decided
			} else{
				t = true
			}
			if t && replica.Slots[i].Command.Command != ""{
				log.Printf("[%v] Decide: === Applying Command: %v", req.Slot, replica.Slots[i].Command)
				commands := strings.Fields(replica.Slots[i].Command.Command)
				var decision string
				if commands[0] == "put"{
					replica.Database[commands[1]] = commands[2]
					decision = fmt.Sprintf("put: [%v] set to [%v]", commands[1], commands[2])
				}
				if commands[0] == "get"{
					value := replica.Database[commands[1]]
					decision = fmt.Sprintf("get: [%v] with [%v]", commands[1], value)
				}
				if commands[0] == "delete"{
					value := replica.Database[commands[1]]
					delete(replica.Database, commands[1])
					decision = fmt.Sprintf("delete: [%v] with [%v]", commands[1], value)
				}
				localAddress := getLocalAddress()
				address := (localAddress + ":" + replica.Port)
				log.Printf("Address is %v", address)
				channelkey := address + strconv.Itoa(req.Command.Tag)
				decisionchannel, ok := decisionmap[channelkey]
				if ok {
					decisionchannel <- decision
				}
			}
		}
	}
	time.Sleep(1000 * time.Millisecond)
}

func call(address string, method string, request interface{}, reply interface{}) error {
	client, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		log.Printf("rpc.DialHTTP: %v", err)
		return err
	}

	defer client.Close()

	if err = client.Call(method, request, reply); err != nil{
		log.Printf("client.Call %s: %v", method, err)
		return err
	}

	return nil
}