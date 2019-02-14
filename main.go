package main

import (
	"flag"
	"net"
	"io"
	"strings"
	"log"
	"time"
sc	"strconv"
)


// Global variables
var (
	port		string	// port to listen on
	width	uint64	// width to read from connection
	silo		Silo		// our silo
	tractor	Tractor	// our tractor
	mode	string	// "silo" or "tractor"
	busy		bool	= false	// lock for busy signal
	auth		bool		// auth mode y/n
	pin		int		// pin
)


// Spins for a period of time in minutes
func spin(n int, during, after string) {
	for busy {
		time.Sleep(15 * time.Millisecond)
	}

	busy = true
	silo.status = during
	time.Sleep(time.Duration(n) * time.Minute)
	silo.status = after
	busy = false
}

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
	
	write(banner)	
	
	if auth {
		// Verify auth
		buf := make([]byte, width)
	
		write("Enter PIN: ")
		read(buf)
		
		pinstr := strings.Fields(string(buf))[0]
		
		p, err := sc.Atoi(pinstr)
		
		if err != nil {
			write("Access denied.")
			conn.Close()
			connected = false
		}
		
		if !(p == 0 || p < 0 || p > 999999 || p == pin) {
			write("Access denied.")
			conn.Close()
			connected = false
		} else {
			write("Access granted.")
		}

	}

	if mode == "silo" {
		silo.DoCmd(&connected, write, read, invalid)
	} else if mode == "tractor" {
		tractor.DoCmd(&connected, write, read, invalid)
	}
}

// Simulates a silo, listens on tcp/1337
func main() {
	flag.StringVar(&port, "p", ":1337", "Port to listen on")
	flag.Uint64Var(&width, "w", 1024, "Max width of communications")
	flag.StringVar(&mode, "m", "silo", "Which mode to start in: tractor, silo")
	flag.IntVar(&pin, "c", 1234, "Pin for auth (if any)")
	flag.BoolVar(&auth, "a", true, "Auth t/f")
	flag.Parse()
	
	if mode != "silo" && mode != "tractor" {
		log.Fatal("Error: mode must be one of tractor or silo.")
	}

	if mode == "silo" {
		silo = NewSilo()
	} else if mode == "tractor" {
		tractor = NewTractor()
	}

	// Start listener
	ln, err := net.Listen("tcp", port)
	efatal(err, "couldn't start listener")

	for {
		conn, err := ln.Accept()
		efatal(err, "could not accept connection")

		go handler(conn)
	}
}

// Style: ANSI Shadow
var banner string = `
DON JEERE INDUSTRIES UNIFIED OPERATING SYSTEM
COPYRIGHT 2075-2077 DON JEERE INDUSTRIES
───────────────────────────────────────────────────────────────────────────
DIAL SUCCEEDED
***************************************************************************

██████╗  ██████╗ ███╗   ██╗         ██╗███████╗███████╗██████╗ ███████╗
██╔══██╗██╔═══██╗████╗  ██║         ██║██╔════╝██╔════╝██╔══██╗██╔════╝
██║  ██║██║   ██║██╔██╗ ██║         ██║█████╗  █████╗  ██████╔╝█████╗  
██║  ██║██║   ██║██║╚██╗██║    ██   ██║██╔══╝  ██╔══╝  ██╔══██╗██╔══╝  
██████╔╝╚██████╔╝██║ ╚████║    ╚█████╔╝███████╗███████╗██║  ██║███████╗
╚═════╝  ╚═════╝ ╚═╝  ╚═══╝     ╚════╝ ╚══════╝╚══════╝╚═╝  ╚═╝╚══════╝
`
