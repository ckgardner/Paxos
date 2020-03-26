package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func mainCommands(replica *Replica) {
	log.Printf("Paxos is ready")
	log.Printf("Command options: help, get, put, delete, dump, -chatty=(0-2), -latency=[n, 2n], quit")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		parts := strings.SplitN(line, " ", 3)
		if len(parts) == 0 {
			continue
		}
		//var nothing *Nothing
		switch parts[0] {
			case "help":
				fmt.Println("Usable Commands:")
				fmt.Println("get, put, delete, dump, -chatty=(0-2), -latency=[n, 2n]")

			case "put":

			case "get":

			case "delete":

			case "dump":
				log.Printf("Port:%v", replica.Port)
				log.Printf("Cell:%v", replica.Cell)
				log.Println("Database:")
				for key, value := range replica.Database{
					log.Println(key, "->", value)
				}
				undecided := len(replica.Slots)
				for i := 0; i <len(replica.Slots); i++{
					if !replica.Slots[i].Decided{
						undecided = i
						break
					}
				}
				log.Println("Next undecided slot is:", undecided)
				for slot := range replica.Slots{
					log.Println("slot=", slot)
				}

			case "quit":
				if len(replica.Slots) == 1{
					fmt.Println("Dissolving from slot")
				}
				replica.Kill <- struct{}{}

			default:
				log.Printf("I don't recognize this command")
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("Error in main command loop: %v", err)
		}
	}
	

