package main

import (
	"os"
	"fmt"
	"log"
	//"github.com/mattn/go-sqlite3"
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
	replica.ToApply = -1
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
			replica.portAddress = append(replica.portAddress, replica.IP + ":" + port)
		}
	}
	replica.HighestSlot.HighestN = 0
	replica.HighestSlot.Accepted.SequenceN = -1

	/* 	path := replica.Cell
		options :=
			"?" + "_busy_timout=10000" +
				"&" + "_foreign_keys=ON" +
				"&" + "_locking_mode=NORMAL" +
				"&" + "mode=rw" + 
				"&" + "_synchronous=NORMAL"

		db, err := sql.Open("sqlite3", path + options)
		if err != nil{
			log.Fatalf("Opening DB error: %v", err)
		}
		defer db.Close()

		create_stmt := `
		CREATE TABLE IF NOT EXISTS slots (
			id	integer primary key,
			data text
		`)

		if _, err := db.Exec(create_stmt); err != nil{
			log.Fatalf("running create stmt: %v", err)
		}
		fmt.Println("DB CREATION SUCCESS!")

		slot := &Slot{
			ID: 1,
			Decided: true,
			Command: "Put apple pie",
			Promise: 5,
			
		}

		raw, err := json.MarshalIndent(slot, "", "    ")
		if err != nil{
			log.Fatalf("Encoding slot: %v", err)
		}
		fmt.Println(string(raw))

		//PARSE
		slot2 := new(Slot)
		if err = json.Unmarshal(raw, slot2); err != nil{
			log.Fatalf("decoding json: %v", err)
		}

		// Create all Database values as empty values and then just update from there
		// Select * by id to read all slots in.
	*/

	go func(){
		<-replica.Kill
		os.Exit(0)
	}()
	server(replica)
	mainCommands(replica)
}