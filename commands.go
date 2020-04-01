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
	log.Printf("Type help for a list of commands")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		parts := strings.SplitN(line, " ", 3)
		if len(parts) == 0 {
			continue
		}


		switch parts[0] {
			case "help":
				fmt.Println("Usable Commands:")
				fmt.Println("get, put, delete, dump, quit")

			case "put":
				if len(parts) == 3{

					var item Slot
					item.Accepted.SequenceN = replica.HighestSlot.Accepted.SequenceN 
					item.Command.Command = line
					item.Command.Address = replica.Address
					if err := Propose(replica,item); err != nil{
						log.Println("Did not set key/value pair")
					}else{
						fmt.Println("Complete")
					}

				}else{
					fmt.Println("Put needs <key> <value> pair")
				 }

			case "get":

			case "delete":

			case "dump":
				log.Printf("Port:%v", replica.Port)
				log.Printf("Address:Ports: =============")
				for index := range replica.portAddress{
					fmt.Println("\t"+string(replica.portAddress[index]))
				}

				undecided := replica.HighestSlot.HighestN

				log.Println("Next undecided slot is:", undecided)

				log.Println("Cell: ============")
				for i, value := range replica.Cell{
					log.Printf("\t[%v] -> %v\n", int(i) , value)
				}

			case "quit":
				if len(replica.Cell) == 1{
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
	

