package main

import (
	"os"
	"fmt"
	"log"
)

const (
	defaultHost = "localHost"
	defaultPort = "3410"
)

func main() {
	replica := new(Replica)
	replica.Database = make(map[string]string)
	replica.Kill = make(chan struct{})
	args := os.Args[1:]
	if len(args) < 3{
		log.Fatalln("Needs 3 arguments")
	}
	replica.HighestSlot.HighestN = 0
	replica.HighestSlot.Accepted.SequenceN = -1

	replica.IP = getLocalAddress()
	replica.Port = ":" + args[0]
	replica.Address = string(replica.IP) + string(replica.Port)
	replica.Cell = args
	fmt.Println("My Replica Address: ", replica.Address)

	go func(){
		<-replica.Kill
		os.Exit(0)
	}()
	server(replica)
	mainCommands(replica)
}