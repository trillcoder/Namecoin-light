/*
 * Namecoin-light (Super early beta)
 * By Trillcoder
 *
 * nameserver.go
 *
 * POC Node that provides light peers with namecoin lookups.
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
 * 	go run nameserver.go
 *
 * 	Needs to have namecoind running on the machine.
 *
 *
 * */

package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"log"
	"math"
	"net"
	"os/exec"
	"strconv"
	"strings"
)

type Name struct {
	name               string
	value              string
	txid               string
	address            string
	expires_in         float64
	registration_block float64
}

type Name_history struct {
}

type Nmc_request struct {
	Command string
	Value   string
}

func name_show(name string) string {
	cmd := exec.Command("namecoind", "name_show", name)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		//~ log.Fatal(err)
		return "Error"
	}

	return out.String()

}

func get_command(command string) string {
	//todo extract out the commands
	return command

}

func decode_json(json_string string) map[string]interface{} {

	data := map[string]interface{}{}
	dec := json.NewDecoder(strings.NewReader(json_string))
	dec.Decode(&data)
	return data
}

func (m *Name) load_name(data map[string]interface{}) {
	m.name = data["name"].(string)
	m.value = data["value"].(string)
	m.txid = data["txid"].(string)
	m.address = data["address"].(string)
	m.expires_in = data["expires_in"].(float64)
}

func (m *Name) get_creation_block(current_block float64) {

	m.registration_block = current_block - (36000 - m.expires_in)
}

func (m *Name) get_full_history() {

}

func get_current_block() float64 {
	out, err := exec.Command("namecoind", "getblockcount").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("The date is %s\n", out)
	//convert byte[] to string and also removing all non ascii characters (this also removes whitespace)
	out_string := stripCtlAndExtFromUnicode(string(out[:])) // slice the array from beginning to end

	fmt.Printf("The date is %s\n", out_string)
	out_float, errparse := strconv.ParseFloat(out_string, 64)

	if errparse != nil {
		log.Fatal(errparse)
	}

	return out_float
}

func get_name(arg string, myname *Name) string {

	//still try to make sure that the name is th correct type and sanitize it
	name_json := name_show(arg)
	if name_json == "Error" {
		//~ log.Fatal(err)
		return "Error"
	}

	data := decode_json(name_json)

	myname.load_name(data)

	return myname.value
}

func main() {

	//set connection variables
	counter := 0
	//TODO include peer lookup and peer bootsrap table

	// Listen on TCP port 2000 on all interfaces.
	l, err := net.Listen("tcp", ":1337")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		//Handles the connection in a goroutine
		go func(c net.Conn) {

			dec := gob.NewDecoder(c)
			requested := new(Nmc_request)

			err := dec.Decode(&requested)
			if err != nil {
				log.Fatal("decode error:", err)
			}
			fmt.Println(requested.Value)
			fmt.Println(requested.Command)

			myname := new(Name)
			//SANITIZED with stripCtlAndExtFromUnicode
			command := get_command(stripCtlAndExtFromUnicode(string(requested.Command)))
			if command == "name_show" {

				name := get_name(stripCtlAndExtFromUnicode(string(requested.Value)), myname)
				name_json := name
				fmt.Printf("%s\n", string(requested.Value))
				fmt.Printf("%s\n", string(name_json))
				fmt.Printf("%f\n", float64(myname.expires_in))
				c.Write([]byte(name_json))
				c.Close()

			} else if command == "name_origin" {

				//gets the name
				name_json := get_name(stripCtlAndExtFromUnicode(string(requested.Value)), myname)

				//get the full history
				currentblock := get_current_block()
				fmt.Printf("%s\n", float64(currentblock))

				myname.get_creation_block(currentblock)

				fmt.Printf("Registered on : %s\n", float64(myname.registration_block))

				fmt.Printf("%s\n", string(requested.Value))
				fmt.Printf("%s\n", string(name_json))
				fmt.Printf("%s\n", float64(myname.expires_in))
				c.Write([]byte(name_json))
				c.Close()

			} else {
				name_json := "Command not found"
				fmt.Printf(name_json)
				c.Write([]byte(name_json))
				c.Close()
			}

			fmt.Printf("Counter value : %s\n", int(counter))

			counter = counter + 1

			// Shut down the connection.

		}(conn)
	}

}

//function to sanitize input
//from: http://rosettacode.org/wiki/Strip_control_codes_and_extended_characters_from_a_string#Go
func stripCtlAndExtFromUnicode(str string) string {
	isOk := func(r rune) bool {
		return r < 32 || r >= 127
	}
	// The isOk filter is such that there is no need to chain to norm.NFC
	t := transform.Chain(norm.NFKD, transform.RemoveFunc(isOk))
	// This Transformer could also trivially be applied as an io.Reader
	// or io.Writer filter to automatically do such filtering when reading
	// or writing data anywhere.
	str, _, _ = transform.String(t, str)
	return str
}

//function to convert the byte to float64
//http://stackoverflow.com/questions/22491876/convert-byte-array-uint8-to-float64-in-golang
func Float64frombytes(bytes []byte) float64 {
	bits := binary.LittleEndian.Uint64(bytes)
	float := math.Float64frombits(bits)
	return float
}
