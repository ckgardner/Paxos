package main

import(
    "net"
    "net/rpc"
    "net/http"
    "log"
    "fmt"
)
func server(replica *Replica) {
	location := replica.Port
	rpc.Register(replica)
	rpc.HandleHTTP()
	listener, err := net.Listen("tcp", location)
	if err != nil {
		log.Fatal("Error thrown while listening: ", err)
	}
	fmt.Printf("Listening %v\n", location)

	go func() {
		if err := http.Serve(listener, nil); err != nil {
			log.Fatalf("Serving: %v", err)
		}
	}()
	fmt.Println("Server is on")
}


func getLocalAddress() string {
    var localaddress string

    ifaces, err := net.Interfaces()
    if err != nil {
        panic("init: failed to find network interfaces")
    }

    // find the first non-loopback interface with an IP address
    for _, elt := range ifaces {
        if elt.Flags&net.FlagLoopback == 0 && elt.Flags&net.FlagUp != 0 {
            addrs, err := elt.Addrs()
            if err != nil {
                panic("init: failed to get addresses for network interface")
            }

           for _, addr := range addrs {
                if ipnet, ok := addr.(*net.IPNet); ok {
                    if ip4 := ipnet.IP.To4(); len(ip4) == net.IPv4len {
                        localaddress = ip4.String()
                        break
                    }
                }
            }
        }
    }
    if localaddress == "" {
        panic("init: failed to find non-loopback interface with valid address on this node")
    }

    return localaddress
}