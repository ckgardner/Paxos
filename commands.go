package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"strconv"
	"time"
)

func mainCommands(replica *Replica) {
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
				replica.Kill <- struct{}{}

			default:
				log.Printf("I don't recognize this command")
			}
			if err := scanner.Err(); err != nil {
				fmt.Printf("Error in main command loop: %v", err)
			}
		}
	}
	
