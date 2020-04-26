package main

import (
	"os"
	"fmt"
	"net/http"
	"net/rpc"
	//"github.com/mattn/go-sqlite3"
)

func main() {
	dump = false
	var port string
	var portAddresses []string

	for i, arg := range os.Args {
		if i == 0 {
			continue
		}
		if port == "" {
			port = arg
		} else {
			portAddresses = append(portAddresses, arg)
		}
	}

	replica := &Replica{
		Address:   port,
		Cell:      append(portAddresses, port),
		Slots:     make([]Slot, 100),
		Listeners: make(map[string]chan string),
		Database:  make(map[string]string),
	}

	rpc.Register(replica)
	rpc.HandleHTTP()

	fmt.Printf("New Replica: %v", replica)

	go func() {
		err := http.ListenAndServe(getAddress(port), nil)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}()
	fmt.Println("Server is up and running...")

	if len(os.Args) < 2 {
		fmt.Println("You must provide a port number")
	}else{
		fmt.Println("Main Commands Running...")
		mainCommands(replica)
	}
}