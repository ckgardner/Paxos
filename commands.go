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
		// var nothing *Nothing

		switch parts[0] {
			case "help":
				fmt.Println("Usable Commands:")
				fmt.Println("get, put, delete, dump, quit")

			case "put":
				if len(parts) == 3{

					// pair := Pair{parts[1], parts[2]}

					// if err := call(replica.Address,"Replica.Set", pair, &nothing); err != nil {
					// 	log.Printf("replica.Set: %v", err)
					// }else{

					// 	log.Printf("This was inserted to the Node: {%v:%v}", pair.Key, pair.Value)
					// }

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
				log.Printf("Cell: =============")
				for index := range replica.Cell{
					fmt.Println("\t"+string(replica.Cell[index]))
				}
				log.Println("Database: ============")
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
	

