package main
import(
	"log"
	"time"
	"strings"
	"fmt"
	"strconv"
)
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

// func call(address string, method string, request interface{}, reply interface{}) error {
// 	client, err := rpc.DialHTTP("tcp", address)
// 	if err != nil {
// 		log.Printf("rpc.DialHTTP: %v", err)
// 		return err
// 	}

// 	defer client.Close()

// 	if err = client.Call(method, request, reply); err != nil{
// 		log.Printf("client.Call %s: %v", method, err)
// 		return err
// 	}

// 	return nil
// }