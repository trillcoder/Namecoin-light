/*
 * Namecoin-light (Super early beta)
 * By Trillcoder
 *
 * Client.go
 *
 * POC dials server running the server version.
 * TODO:
 * -ADD P2P so that the system works decentralized
 * -ADD bloom filter support. This helps add an SPV like cryptographic
 * 	proof That the value for the lookup is the correct one
 * -Add global DNS service. Allow clients to transparently use
 * 	A light resolver without having to do more than install the exe
 *
 *
 * 	Remeber this is a work in progress and I am still working on it.
 * 	Alot more needs to be done.
 *
 * 	Usage :
 * 	go run client.go name_show d/wikileaks
 *
 * (also supports the other id/ a/ etc. namespaces)
 *
 * 	Needs to have namecoind running on the machine. That won't be
 * 	needed when the p2p implementation is up
 *
 *
 * */

package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"log"
	"net"
	"os"
)

//request structure

type Nmc_request struct {
	Command string
	Value   string
}

func main() {

	//intialize the struct
	args := new(Nmc_request)

	//just adding some simple error checking
	//when i tried to run this program blindly i just got a fail
	//should do some additional testing for acceptable commands/values
	if(len(os.Args) != 3){
		log.Fatal("Requires two arguments: Command and Value, respectively")
	}
	
	args.Command = os.Args[1]
	args.Value = os.Args[2]

	conn, err := net.Dial("tcp", "localhost:1337")
	if err != nil {
		// handle error
		log.Fatal("encode error:", err)
	}

	enc := gob.NewEncoder(conn)
	err = enc.Encode(args)
	if err != nil {
		log.Fatal("encode error:", err)
	}

	status, err := bufio.NewReader(conn).ReadString('\n')

	fmt.Println(status)

}
