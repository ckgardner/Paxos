package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/rpc"
	"strconv"
	"strings"
	"time"
)

var lag int
var dump bool

func (r Command) String() string {
	return fmt.Sprintf("Command:\n\tUniqueRandomTag: %d\n\tCommand: %s\n\t%s\n\tAddress: %s\n", r.Tag, r.Command, r.Sequence.String(), r.Address)
}
func (elt Sequence) String() string {
	return fmt.Sprintf("Sequence:\n\tN: %d\n\tAddress: %s\n", elt.Number, elt.Address)
}
func (slot Slot) String() string {
	return fmt.Sprintf("Slot:\n\tCommand\n\t %v, Decided: %t\n", slot.Command, slot.Decided)
}

// Propose function
func (replica *Replica) Propose(req AllRequests, res *Response) error {
	lagTime := float64(lag)
	timeOffset := float64(lagTime) * rand.Float64()
	time.Sleep(time.Second * time.Duration(lagTime*timeOffset))
	if dump {
		var dumpRes Response
		var dumpReq AllRequests
		err := call(replica.Address, "Dump", dumpReq, &dumpRes)
		if err != nil {
			log.Fatal("Dump Err")
		}
	}

	proposeReq := req.Propose
	fmt.Println("Propose: Command: " + proposeReq.Command.String())

	var accept Accept
	round := 1

	var slot Slot
	slot.Sequence = proposeReq.Command.Sequence
	if slot.Sequence.Number == 0 {
		slot.Sequence.Address = replica.Address
		slot.Sequence.Number = 1
	}
	slot.Command = proposeReq.Command
	slot.Command.Sequence = slot.Sequence

votingRounds:
	for {
		for i := 1; i < len(replica.Slots); i++ {
			if !replica.Slots[i].Decided {
				slot.Number = i
				break
			}
		}
		accept.Slot = slot

		fmt.Println("Propose: Round: " + strconv.Itoa(round))
		fmt.Println("Propose: Slot: " + strconv.Itoa(slot.Number))
		fmt.Println("Propose: N: " + strconv.Itoa(slot.Sequence.Number))

		resp := make(chan Response, len(replica.Cell))
		for _, i := range replica.Cell {
			go func(i string, slot Slot, seq Sequence, resp chan Response) {
				request1 := AllRequests{Address: replica.Address, Prepare: PrepareRequest{Slot: slot, Sequence: seq}}
				var response1 Response
				err := call(i, "Prepare", request1, &response1)
				if err != nil {
					log.Fatal("Prepare Failed", err)
					return
				}
				resp <- response1
			}(i, slot, slot.Sequence, resp)
		}

		rTrue := 0
		rFalse := 0
		highestV := 0
		var highestCommand Command
		for votes := 0; votes < len(replica.Cell); votes++ {
			prepResponse := <-resp
			if prepResponse.IsOkay {
				rTrue++
			} else {
				rFalse++
			}

			if prepResponse.Promise.Number > highestV {
				highestV = prepResponse.Promise.Number
			}
			if prepResponse.Command.Sequence.Number > highestCommand.Sequence.Number {
				highestCommand = prepResponse.Command
			}
			if rTrue >= haveMajority(len(replica.Cell)) || rFalse >= haveMajority(len(replica.Cell)) {
				break
			}
		}

		if rTrue >= haveMajority(len(replica.Cell)) {
			accept.Command = slot.Command
			accept.Seq = slot.Sequence

			// fmt.Print(highestCommand)
			// fmt.Print()
			if highestCommand.Tag > 0 && highestCommand.Tag != proposeReq.Command.Tag {
				accept.Command = highestCommand
				accept.Slot.Command = highestCommand

				req.Accept = accept
				call(replica.Address, "Accept", req, res)
			} else {
				break votingRounds
			}
		}

		slot.Sequence.Number = highestV + 1
		round++
	}
	req.Accept = accept
	call(replica.Address, "Accept", req, res)
	return nil
}

// Prepare is not an RPC
func (replica *Replica) Prepare(req AllRequests, res *Response) error {

	replica.Lock.Lock()
	defer replica.Lock.Unlock()

	time.Sleep(1000 * time.Millisecond)
	log.Println("Prepare called with:", req.Prepare)

	if dump {
		var dumpRes Response
		var dumpReq AllRequests
		err := call(replica.Address, "Dump", dumpReq, &dumpRes)
		if err != nil {
			log.Fatal("Dump Err")
		}
	}

	args := req.Prepare

	if replica.Slots[args.Slot.Number].Sequence.Cmp(args.Sequence) == -1 {
		fmt.Println("Prepare Okay")
		res.IsOkay = true
		res.Promise = args.Sequence
		replica.Slots[args.Slot.Number].Sequence = args.Sequence
		res.Command = replica.Slots[args.Slot.Number].Command

	} else {
		res.IsOkay = false
		fmt.Println("Prepare failed: Already promised higher number")
		res.Promise = replica.Slots[args.Slot.Number].Sequence
	}

	time.Sleep(1000 * time.Millisecond)
	return nil
}

// Accept is not an RPC
func (replica *Replica) Accept(req AllRequests, res *Response) error {
	// Latency sleep
	lagTime := float64(lag)
	timeOffset := float64(lagTime) * rand.Float64()
	time.Sleep(time.Second * time.Duration(lagTime*timeOffset))
	if dump {
		var dumpRes Response
		var dumpReq AllRequests
		err := call(replica.Address, "Dump", dumpReq, &dumpRes)
		if err != nil {
			log.Fatal("Dump Err")
		}
	}

	acceptReq := req.Accept
	aSlot := acceptReq.Slot
	aN := acceptReq.Seq
	aV := acceptReq.Command

	response := make(chan Response, len(replica.Cell))
	for _, v := range replica.Cell {
		go func(v string, slot Slot, sequence Sequence, command Command, response chan Response) {
			req := AllRequests{Address: replica.Address, Accepted: AcceptRequest{Slot: slot, Seq: sequence, Command: command}}
			var resp Response
			err := call(v, "Accepted", req, &resp)
			if err != nil {
				fmt.Println("Error in Accept")
				return
			}
			response <- resp

		}(v, aSlot, aN, aV, response)
	}

	nTrue := 0
	nFalse := 0
	highestN := 0
	for numVotes := 0; numVotes < len(replica.Cell); numVotes++ {
		prepareResponse := <-response
		if prepareResponse.IsOkay {
			nTrue++
		} else {
			nFalse++
		}

		if prepareResponse.Promise.Number > highestN {
			highestN = prepareResponse.Promise.Number
		}

		if nTrue >= haveMajority(len(replica.Cell)) || nFalse >= haveMajority(len(replica.Cell)) {
			break
		}
	}
	time.Sleep(1000 * time.Millisecond)

	if nTrue >= haveMajority(len(replica.Cell)) {
		fmt.Println("Received enough votes!")
		for _, v := range replica.Cell {
			go func(v string, slot Slot, command Command) {
				req := AllRequests{Address: replica.Address, Decide: DecideRequest{Slot: slot, Value: command}}
				var resp Response
				err := call(v, "Decide", req, &resp)
				if err != nil {
					log.Fatal("Decide (from Accept)")
					return
				}
			}(v, aSlot, aV)
		}

		return nil
	}

	fmt.Println("Not enough votes!")
	aV.Sequence.Number = highestN + 1

	duration := float64(5)
	offset := float64(duration) * rand.Float64()
	time.Sleep(time.Millisecond * time.Duration(duration+offset))

	req1 := AllRequests{Address: replica.Address, Propose: ProposeRequest{Command: aV}}
	var resp1 Response
	err := call(replica.Address, "Propose", req1, &resp1)
	if err != nil {
		log.Fatal("Propose")
		return err
	}

	return nil
}

// Accepted is not an RPC
func (replica *Replica) Accepted(req AllRequests, resp *Response) error {
	replica.Lock.Lock()
	defer replica.Lock.Unlock()
	log.Println("=====", replica.Address, "Accepting..")
	if dump {
		var dumpRes Response
		var dumpReq AllRequests
		err := call(replica.Address, "Dump", dumpReq, &dumpRes)
		if err != nil {
			log.Fatal("Dump Err")
		}
	}

	args := req.Accepted

	if replica.Slots[args.Slot.Number].Sequence.Cmp(args.Seq) == 0 {
		fmt.Println("Accepted:  Sequence YES")
		resp.IsOkay = true
		resp.Promise = replica.Slots[args.Slot.Number].Sequence
		replica.ToApply = args.Slot
	} else {
		fmt.Println("Accepted:  Sequence NO. Already promised a higher number: " + strconv.Itoa(replica.Slots[args.Slot.Number].Sequence.Number))
		resp.IsOkay = false
		resp.Promise = replica.ToApply.Sequence
	}
	return nil
}

// Decide is not an  RPC
func (replica *Replica) Decide(req AllRequests, resp *Response) error {

	replica.Lock.Lock()
	defer replica.Lock.Unlock()
	time.Sleep(1000 * time.Millisecond)

	if dump {
		var dumpRes Response
		var dumpReq AllRequests
		err := call(replica.Address, "Dump", dumpReq, &dumpRes)
		if err != nil {
			log.Fatal("Dump Err")
		}
	}

	request := req.Decide

	if replica.Slots[request.Slot.Number].Decided && replica.Slots[request.Slot.Number].Command.Command != request.Value.Command {
		fmt.Println("Decide:  Already decided slot " + strconv.Itoa(request.Slot.Number) + " with a different command " + request.Value.Command)
		log.Fatal("Decide")
		return nil
	}
	// If already decided, quit
	if replica.Slots[request.Slot.Number].Decided {
		fmt.Println("> Decide:  Already decided slot " + strconv.Itoa(request.Slot.Number) + " with command " + request.Value.Command)
		return nil
	}

	_, ok := replica.Listeners[request.Value.Key]
	if ok {
		replica.Listeners[request.Value.Key] <- request.Value.Command
	}

	command := strings.Split(request.Value.Command, " ")
	request.Slot.Decided = true
	replica.Slots[request.Slot.Number] = request.Slot

	fmt.Println(request.Slot.Number)
	for !Decided(replica.Slots, request.Slot.Number) {
		time.Sleep(time.Second)
	}
	switch command[0] {
	case "put":
		log.Println("Put " + command[1] + " " + command[2])
		replica.Database[command[1]] = command[2]
		break
	case "get":
		log.Println("Get{" + command[1] + "," + replica.Database[command[1]] + "}")
		break
	case "delete":
		log.Println(command[1] + " deleted.")
		delete(replica.Database, command[1])
		break
	}

	time.Sleep(1000 * time.Millisecond)
	return nil
}

func call(address, method string, request AllRequests, reply *Response) error {
	client, err := rpc.DialHTTP("tcp", getAddress(address))
	if err != nil {
		log.Printf("rpc.DialHTTP: %v", err)
		return err
	}

	defer client.Close()

	if err = client.Call("Replica."+method, request, reply); err != nil {
		log.Printf("client.Call %s: %v", method, err)
		return err
	}
	return nil
}

func haveMajority(group int) int {
	return int(math.Ceil(float64(group) / 2))
}

// Cmp not RPC
func (elt Sequence) Cmp(rhs Sequence) int {
	if elt.Number == rhs.Number {
		if elt.Address > rhs.Address {
			return 1
		}
		if elt.Address < rhs.Address {
			return -1
		}
		return 0
	}
	if elt.Number < rhs.Number {
		return -1
	}
	if elt.Number > rhs.Number {
		return 1
	}
	return 0
}

// Decided not and RPC
func Decided(slots []Slot, number int) bool {
	if number == 1 {
		return true
	}
	for i, j := range slots {
		if i == 0 {
			continue
		}
		if i >= number {
			break
		}
		if !j.Decided {
			return false
		}
	}
	return true
}

// Dump is not an RPC call
func (replica *Replica) Dump(_ AllRequests, _ *Response) error {
	log.Println("Data")
	for k, v := range replica.Database {
		fmt.Printf("\tKey: %s, Value: %s\n", k, v)
	}
	log.Println("Cell")
	for k, v := range replica.Slots {
		if v.Decided {
			fmt.Printf("\t{%d,%v}\n", k, v)
		}
	}
	return nil
}
