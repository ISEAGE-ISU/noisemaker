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
	"encoding/json"
)


// Global variables
var (
	port		string			// port to listen on
	width		uint64			// width to read from connection
	silo		Silo			// our silo
	tractor		Tractor			// our tractor
	combine 	Combine			// our combine
	mode		string			// "silo", "tractor", or "combine"
	busy		bool = false	// lock for busy signal
	auth		bool			// auth mode y/n
	pin			int				// pin
	status		string			// current status
	reqChan		chan string		// status request channel, unbuffered
	statChan	chan string		// status channel, unbuffered
	dumpChan	chan map[string]string		// dump trigger channel
)

var path = os.Getenv("HOME") + "/farmstate.json"

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

// Manages status get/set
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
			time.Sleep(5 * time.Millisecond)
		}
	}
}

// Manages state writing to file
func dumper() {
	for {
		select {
		case cfg := <- dumpChan:
			f, err := os.Create(path)
			if err != nil {
				log.Println("Error: unable to create file --", err)
				break
			}
			enc := json.NewEncoder(f)
			enc.Encode(cfg)
		default:
			time.Sleep(5 * time.Millisecond)
		}
	}
}

// Handles incoming connections
func handler(conn net.Conn) {
	connected := true
	msgChan := make(chan string)

	defer conn.Close()
	defer log.Println("Handler ended:", conn.RemoteAddr())
	
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

		// Debug flag stuff
		if pinstr == "F" {
			switch mode {
			case "silo":
				write(silo.flag)
			case "tractor":
				write(tractor.flag)
			case "combine":
				// TODO
				;
			}
		}
		
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

		n, err := conn.Read(buf)
		if err != nil || n <= 0 {
			close(msgChan)
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
	reqChan = make(chan string)
	statChan = make(chan string)
	go statuser()
	
	// Load config file
	var cfg map[string]string
	f, err := os.Open(path)
	if err != nil {
		log.Println("Warning: failed to open config file --", err.Error())
	} else {
		dec := json.NewDecoder(f)
		err := dec.Decode(&cfg)
		if err != nil {
			log.Println("Warning: config file decoding failed --", err.Error())
		}
	}
	
	dumpChan = make(chan map[string]string, 5)
	go dumper()

	switch mode {
	case "silo":
		silo = NewSilo()
		silo.LoadCfg(cfg)
	case "tractor":
		tractor = NewTractor()
		tractor.LoadCfg(cfg)
	case "combine":
		combine = NewCombine()
		// TODO -- loadcfg
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
GEN 2.2
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
