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
	replica.HighestSlot.HighestN = 0
	replica.ToApply = 0
	replica.HighestSlot.Command.SequenceN = -1
	args := os.Args
	args = os.Args[1:]
	if len(args) < 3{
		log.Fatalln("Needs 3 arguments")
	}else{

		replica.IP = getLocalAddress()
		replica.Port = ":" + args[0]
		replica.Address = string(replica.IP) + string(replica.Port)
		// replica.Cell = args
		fmt.Println("My Replica Address: ", replica.Address)
		for _,port := range args{
			replica.Cell = append(replica.Cell, replica.IP + ":" + port)
		}
	}



	go func(){
		<-replica.Kill
		os.Exit(0)
	}()
	server(replica)
	mainCommands(replica)
}