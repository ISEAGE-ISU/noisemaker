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
			write(silo.cont)
		
		// Supply
		case "supply":
			switch len(argv) {
			case 1:
				write(sc.Itoa(silo.supply))
			case 2:
				if argv[1] == "load" {
					// TODO ­ more max/min logic
					silo.supply += 10
				} else {
					// unload
					silo.supply -= 10
				}
			default:
				invalid()
			}
		
		// Heat
		case "heat":
			switch len(argv) {
			case 1:
				write(sc.Itoa(silo.temp))
			case 3:
				n, err := sc.Atoi(argv[2])

				if err != nil {
					invalid()
				}

				if argv[1] == "raise" {
					// TODO ­ more max/min logic
					silo.temp += n
				} else {
					// lower
					silo.temp -= n
				}
			default:
				invalid()
			}
		
		// Humidity
		case "humidity":
			switch len(argv) {
			case 1:
				write(sc.Itoa(silo.humid))
			case 3:
				n, err := sc.Atoi(argv[2])

				if err != nil {
					invalid()
				}

				if argv[1] == "raise" {
					// TODO ­ more max/min logic
					silo.humid += n
				} else {
					// lower
					silo.humid -= n
				}
			default:
				invalid()
			}
		
		// Status
		case "status":
			write(silo.status)
		
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
	silo.status	= "off"
	silo.humid	= 30
	silo.temp	= 20
	silo.supply	= 0
	silo.cont	= "corn"
	silo.pin	= 1234
	silo.auth	= true

	// Start listener
	ln, err := net.Listen("tcp", port)
	efatal(err, "couldn't start listener")

	for {
		conn, err := ln.Accept()
		efatal(err, "could not accept connection")

		go handler(conn)
	}
}

