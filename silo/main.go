package main

import (
	"flag"
	"net"
	"io"
	"strings"
	"log"
)


// Global variables
var (
	port	string	// port to listen on
	width	uint64	// width to read from connection
	silo	Silo	// our silo
)


// Handles incoming connections
func handler(conn net.Conn) {
	defer conn.Close()
	
	// Handles basic writing to interface
	write := func(msg string) {
		_, err := conn.Write([]byte(msg + "\n"))
		if err != nil {
			log.Println("Failed to write, disconnecting: ", err)
			conn.Close()
		}
	}
	
	invalid := func() {
		write("err: invalid arguments")
	}

	for {
		buf := make([]byte, width)

		_, err := conn.Read(buf)

		if err != nil {
			if err == io.EOF {
				// Connection ended
				break
			} else {
				log.Println("Unexpected disconnect: ", err)
				break
			}
		}

		argv := strings.Fields(string(buf))
		
		if len(argv) > 1 {
			// '\n' counts as a field split, truncate
			argv = argv[:len(argv)-1]
		}

		if len(argv) < 1 {
			write("err: no command specified")
		}

		switch argv[0] {
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
		
		// Manual disconnect commands, for convenience
		case "quit":
			fallthrough
		case "exit":
			write("ok")
			conn.Close()
			break

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

	// Start listener
	ln, err := net.Listen("tcp", port)
	efatal(err, "couldn't start listener")

	for {
		conn, err := ln.Accept()
		efatal(err, "could not accept connection")

		go handler(conn)
	}
}

