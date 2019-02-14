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
	tractor	Tractor	// our tractor
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
	
	
	if tractor.auth {
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
		
		if !(pin == 0 || pin < 0 || pin > 999999 || pin == tractor.pin) {
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
		
		if !tractor.power && argv[0] != "power" {
			write("err: powered off")
			goto nocmd
		}

		// Commands master switch
		switch argv[0] {
		
		// Lights
		case "lights":
			switch len(argv) {
			case 1:
				write(string(tractor.Lights()))
			case 2:
				if argv[1] == "on" {
					tractor.lights = true
				} else {
					tractor.lights = false
				}
			default:
				invalid()
			}
			
		// Oil
		case "oil":
			switch len(argv) {
			case 1:
				write(sc.Itoa(tractor.oil))
			case 2:
				if argv[1] == "add" {
					tractor.oil += 100
				} else {
					invalid()
				}
			default:
				invalid()
			}
		
		// Fuel
		case "fuel":
			switch len(argv) {
			case 1:
				write(sc.Itoa(tractor.fuel))
			case 2:
				if argv[1] == "add" {
					tractor.fuel += 100
				} else {
					tractor.fuel = 0
				}
			default:
				invalid()
			}
			
		// Tires
		case "tires":
			switch len(argv) {
			case 1:
				write(sc.Itoa(tractor.tires))
			case 2:
				if argv[1] == "inflate" {
					tractor.tires = 100
				} else {
					tractor.tires = 0
				}
			default:
				invalid()
			}
		
		// Harvest
		case "harvest":
			switch len(argv) {
				case 1:
					write(tractor.status)
				case 2:
					if argv[1] == "start" {
						tractor.status = "harvesting"
					} else {
						tractor.status = "idle"
					}
				default:
					invalid()
				}

		// Contents
		case "contents":
			write(tractor.cont)
		
		// Power
		case "power":
			switch len(argv) {
			case 1:
				write(string(tractor.Power()))
			case 2:
				if argv[1] == "on" {
					tractor.power = true
				} else {
					tractor.power = false
				}
			default:
				invalid()
			}
		
		// Supply
		case "supply":
			switch len(argv) {
			case 1:
				write(sc.Itoa(tractor.supply))
			case 3:
				n, err := sc.Atoi(argv[2])

				if err != nil {
					invalid()
				}

				if argv[1] == "load" {
					// TODO ­ more max/min logic
					tractor.supply += n
				} else {
					// lower
					tractor.supply -= n
				}
			default:
				invalid()
			}
		
		// Heat
		case "heat":
			switch len(argv) {
			case 1:
				write(sc.Itoa(tractor.temp))
			case 3:
				n, err := sc.Atoi(argv[2])

				if err != nil {
					invalid()
				}

				if argv[1] == "raise" {
					// TODO ­ more max/min logic
					tractor.temp += n
				} else {
					// lower
					tractor.temp -= n
				}
			default:
				invalid()
			}
		
		// Humidity
		case "humidity":
			switch len(argv) {
			case 1:
				write(sc.Itoa(tractor.humid))
			case 3:
				n, err := sc.Atoi(argv[2])

				if err != nil {
					invalid()
				}

				if argv[1] == "raise" {
					// TODO ­ more max/min logic
					tractor.humid += n
				} else {
					// lower
					tractor.humid -= n
				}
			default:
				invalid()
			}
		
		// Status
		case "status":
			write(tractor.status)
		
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
		nocmd:
	}
}

// Simulates a tractor, listens on tcp/1337
func main() {
	flag.StringVar(&port, "p", ":1337", "Port to listen on")
	flag.Uint64Var(&width, "w", 1024, "Max width of communications")
	flag.Parse()
	
	// Init tractor
	tractor.status	= "idle"
	tractor.humid	= 30
	tractor.temp	= 20
	tractor.supply	= 0
	tractor.cont	= "corn"
	tractor.pin	= 1234
	tractor.auth	= true
	tractor.fuel 	= 0
	tractor.oil		= 0
	tractor.tires	= 100

	// Start listener
	ln, err := net.Listen("tcp", port)
	efatal(err, "couldn't start listener")

	for {
		conn, err := ln.Accept()
		efatal(err, "could not accept connection")

		go handler(conn)
	}
}
