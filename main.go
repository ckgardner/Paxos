package main

import (
	"os"
	"fmt"
	"net/http"
	"net/rpc"
	"strings"
	"strconv"
	"log"
	"bufio"
	"time"
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
		log.Printf("Paxos is ready")
		log.Printf("Type help for a list of commands")

		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			userCommand := strings.Split(scanner.Text(), " ")
			switch userCommand[0] {
				case "help":
					fmt.Println("Usable Commands:")
					fmt.Println("get, put, delete, dumpall, quit")

				case "put":
					var command Command
					command.Command = strings.Join(userCommand, " ")
					command.Address = replica.Address
					command.Tag, _ = strconv.ParseInt(strconv.FormatInt(time.Now().Unix(), 10)+command.Address, 10, 64)
					respChan := make(chan string, 1)
					key := strconv.FormatInt(command.Tag, 10) + command.Address
					command.Key = key
					replica.Listeners[key] = respChan

					req := AllRequests{Address: replica.Address, Propose: ProposeRequest{Command: command}}
					var resp Response
					err := call(replica.Address, "Propose", req, &resp)
					if err != nil {
						log.Fatal("Propose")
						continue
					}

					go func() {
						fmt.Println("Finished " + <-replica.Listeners[key])
					}()
					break

				case "get":
					var command Command
					command.Command = strings.Join(userCommand, " ")
					command.Address = replica.Address
					// Assign the command a tag
					command.Tag, _ = strconv.ParseInt(strconv.FormatInt(time.Now().Unix(), 10)+command.Address, 10, 64)
					// create a string channel with capacity 1 where the response to the command can be communicated back to the shell code that issued the command
					respChan := make(chan string, 1)
					// store the channel in a map associated with the entire replica. it should map the address and tag number (combined into a string) to the channel
					key := strconv.FormatInt(command.Tag, 10) + command.Address
					command.Key = key
					replica.Listeners[key] = respChan

					req := AllRequests{Address: replica.Address, Propose: ProposeRequest{Command: command}}
					var resp Response
					err := call(replica.Address, "Propose", req, &resp)
					if err != nil {
						log.Fatal("Propose")
						continue
					}

					go func() {
						fmt.Println("Finished " + <-replica.Listeners[key])
					}()
					break

				case "delete":
					var command Command
					command.Command = strings.Join(userCommand, " ")
					command.Address = replica.Address
					command.Tag, _ = strconv.ParseInt(strconv.FormatInt(time.Now().Unix(), 10)+command.Address, 10, 64)
					respChan := make(chan string, 1)
					key := strconv.FormatInt(command.Tag, 10) + command.Address
					command.Key = key
					replica.Listeners[key] = respChan

					req := AllRequests{Address: replica.Address, Propose: ProposeRequest{Command: command}}
					var resp Response
					err := call(replica.Address, "Propose", req, &resp)
					if err != nil {
						log.Fatal("Propose")
						continue
					}

					go func() {
						log.Println("Finished " + <-replica.Listeners[key])
					}()
					break

				case "dump":
					for _, v := range replica.Cell {
						var resp Response
						var req AllRequests
						err := call(v, "Dump", req, &resp)
						if err != nil {
							log.Fatal("dump")
							continue
						}
					}
					break

				case "quit":
					if len(replica.Cell) == 1{
						fmt.Println("Dissolving from slot")
					}
					fmt.Println("Quit Node")
					os.Exit(3)

				default:
					log.Printf("I don't recognize this command")

				if err := scanner.Err(); err != nil {
					fmt.Printf("Error in main command loop: %v", err)
				}
			}
		}
	}
}