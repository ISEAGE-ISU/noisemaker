package main

import (
	"flag"
	"net"
	"io"
	"strings"
	"log"
sc	"strconv"
)


// Global variables
var (
	port	string	// port to listen on
	width	uint64	// width to read from connection
	silo	Silo	// our silo
)


// Handles incoming connections
func handler(conn net.Conn) {
	connected := true

	defer conn.Close()
	
	// Handles basic writing to interface
	write := func(msg string) {
		_, err := conn.Write([]byte(msg + "\n"))
		if err != nil {
			log.Println("Failed to write, disconnecting: ", err)
			conn.Close()
			connected = false
		}
	}
	
	// Handles basic reading to interface
	read := func(buf []byte) {
		_, err := conn.Read(buf)

		if err != nil {
			if err == io.EOF {
				// Connection ended
				conn.Close()
				connected = false
			} else {
				log.Println("Unexpected disconnect: ", err)
				conn.Close()
				connected = false
			}
		}
	}
	
	invalid := func() {
		write("err: invalid arguments")
	}
	
	
	if silo.auth {
		// Verify auth
		buf := make([]byte, width)
	
		write("Enter PIN: ")
		read(buf)
		
		pinstr := strings.Fields(string(buf))[0]
		
		pin, err := sc.Atoi(pinstr)
		
		if err != nil {
			write("Access denied.")
			conn.Close()
			connected = false
		}
		
		if !(pin == 0 || pin < 0 || pin > 999999 || pin == silo.pin) {
			write("Access denied.")
			conn.Close()
			connected = false
		} else {
			write("Access granted.")
		}

	}

	for connected {
		buf := make([]byte, width)

		read(buf)

		argv := strings.Fields(string(buf))
		
		if len(argv) > 1 {
			// '\n' counts as a field split, truncate
			argv = argv[:len(argv)-1]
		}

		if len(argv) < 1 {
			write("err: no command specified")
		}

		// Commands master switch
		switch argv[0] {
		
		// Lights
		case "lights":
			log.Println(argv, len(argv))

			switch len(argv) {
			case 1:
				write(string(silo.Lights()))
			case 2:
				if argv[1] == "on" {
					silo.lights = true
				} else {
					silo.lights = false
				}
			default:
				invalid()
			}
		
		
		// Contents
		case "contents":
		
		
		// Supply
		case "supply":
		
		
		// Heat
		case "heat":
		
		
		// Humidity
		case "humidity":
		
		
		// Status
		case "status":
		
		
		// Manual disconnect commands, for convenience
		case "quit":
			fallthrough
		case "exit":
			write("ok")
			conn.Close()
			break

		// Command not found
		default:
			write("err: unknown command")
		}
		
	}
}

// Simulates a silo, listens on tcp/1337
func main() {
	flag.StringVar(&port, "p", ":1337", "Port to listen on")
	flag.Uint64Var(&width, "w", 1024, "Max width of communications")
	flag.Parse()
	
	// Init silo
	silo.pin = 1234
	silo.auth = true

	// Start listener
	ln, err := net.Listen("tcp", port)
	efatal(err, "couldn't start listener")

	for {
		conn, err := ln.Accept()
		efatal(err, "could not accept connection")

		go handler(conn)
	}
}

