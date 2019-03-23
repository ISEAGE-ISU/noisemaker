package main

import (
	"flag"
	"net"
	"io"
	"strings"
	"log"
	"time"
sc	"strconv"
	"os"
)


// Global variables
var (
	port		string	// port to listen on
	width	uint64	// width to read from connection
	silo		Silo		// our silo
	tractor	Tractor	// our tractor
	combine Combine // our combine
	mode	string	// "silo", "tractor", or "combine"
	busy		bool	= false	// lock for busy signal
	auth		bool		// auth mode y/n
	pin		int		// pin
	status	string	// current status
	reqChan		chan string	// status request channel, unbuffered
	statChan	chan string	// status channel, unbuffered
)


// Spins for a period of time in minutes
func spin(n int, during, after string) {
	for busy {
		time.Sleep(15 * time.Millisecond)
	}

	// Might be bad
	busy = true
	
	reqChan <- during
	
	time.Sleep(time.Duration(n) * time.Minute)

	reqChan <- after
	
	busy = false
}

// Handshakes for current status
func stat() string {
	reqChan <- ""
	return <-statChan
}

// Manages status
func statuser() {
	for {
		select {
		case req := <- reqChan:
			switch req {
			case "":
				// Ordinary request
				statChan <- status
			default:
				status = req
			}
		default:
			time.Sleep(5)
		}
	}
}

// Handles incoming connections
func handler(conn net.Conn) {
	connected := true
	msgChan := make(chan string)

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
	
	// "Do it with flare" ­ Clockwerk
	write(banner)	
	
	// PIN Authentication
	if auth {
		// Verify auth
		buf := make([]byte, width)
	
		write("Enter PIN: ")
		read(buf)
		
		pinstr := strings.Fields(string(buf))[0]
		
		p, err := sc.Atoi(pinstr)
		
		if !(p == 0 || p < 0 || p > 999999 || p == pin) || err != nil {
			write("Access denied.")
			conn.Close()
			return
		} else {
			write("Access granted.\n")
		}

	}

	switch mode {
	case "silo":
		go silo.DoCmd(msgChan)
	case "tractor":
		go tractor.DoCmd(msgChan)
	case "combine":
		write("Vendor must add Combine support for 2.0 control schema")
		os.Exit(1337)
	}
	
	// Empty strings are indicative of desiring teardown
	for connected {
		// Errors can be dealt with later
		conn.Write([]byte("> "))
	
		buf := make([]byte, width)

		_, err := conn.Read(buf)
		if err != nil {
			break
		}

		msgChan <- string(buf)
		
		msg, more := <- msgChan

		// If quit or similar command to end connection
		if !more {
			connected = false
			msg = "ok."
		}
		
		write(msg)
	}

	if _, more := <- msgChan; more {
		close(msgChan)
	}

	log.Println("Handler ended")
}

// Simulates a silo, listens on tcp/1337
func main() {
	flag.StringVar(&port, "p", ":1337", "Port to listen on")
	flag.Uint64Var(&width, "w", 1024, "Max width of communications")
	flag.StringVar(&mode, "m", "combine", "Which mode to start in: tractor, silo, combine")
	flag.IntVar(&pin, "c", 1234, "Pin for auth (if any)")
	flag.BoolVar(&auth, "a", true, "Auth t/f")
	flag.Parse()
	
	if mode != "silo" && mode != "tractor" && mode != "combine" {
		log.Fatal("Error: mode must be one of tractor, silo, or combine.")
	}

	// Set up status management
	status = "idle"
	statChan = make(chan string)
	reqChan = make(chan string)
	go statuser()

	switch mode {
	case "silo":
		silo = NewSilo()
	case "tractor":
		tractor = NewTractor()
	case "combine":
		combine = NewCombine()
	default:
		log.Fatal("Error: Invalid mode --", mode)
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
TRAC CORP INDUSTRIES UNIFIED OPERATING SYSTEM
COPYRIGHT 2075-2077 TRAC CORP INDUSTRIES
───────────────────────────────────────────────────────────────────────────
DIAL SUCCEEDED
***************************************************************************

████████╗██████╗  █████╗  ██████╗     ██████╗ ██████╗ ██████╗ ██████╗ 
╚══██╔══╝██╔══██╗██╔══██╗██╔════╝    ██╔════╝██╔═══██╗██╔══██╗██╔══██╗
   ██║   ██████╔╝███████║██║         ██║     ██║   ██║██████╔╝██████╔╝
   ██║   ██╔══██╗██╔══██║██║         ██║     ██║   ██║██╔══██╗██╔═══╝ 
   ██║   ██║  ██║██║  ██║╚██████╗    ╚██████╗╚██████╔╝██║  ██║██║     
   ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝     ╚═════╝ ╚═════╝ ╚═╝  ╚═╝╚═╝     
`
