package main

import (
	"strings"
	"log"
sc	"strconv"
)

// Silo type
type Silo struct {
	lights	bool	// whether silo lights are on or not
	power	bool	// power on/off
	humid	int		// current humidity %
	temp	int		// current temp °C
	supply	int		// current supply levels bushels
	cont	string	// current contents
	flag	string	// flag string
}

// Printable lights format
func (s *Silo) Lights() string {
	if s.lights {
		return "on"
	}
	
	return "off"
}

// Printable lights format
func (s *Silo) Power() string {
	return b2s(s.power)
}

// Constructor
func NewSilo() (s Silo) {
	// Init silo
	s.humid	= 30
	s.temp	= 20
	s.supply	= 0
	s.cont	= "corn"

	return
}

func (s *Silo) DoCmd(msgChan chan string) {
	ok := func() {msgChan <- "ok."}
	invalid := func() {msgChan <- "err: invalid arguments"}

	for {
		buf, more := <- msgChan
		if !more {
			break 
		}
	
		argv := strings.Fields(buf)
		
		if len(argv) > 1 {
			// '\n' counts as a field split, truncate
			argv = argv[:len(argv)-1]
		}

		if len(argv) < 1 {
			msgChan <- "err: no command specified"
			continue
		}
		
		if !s.power && argv[0] != "power" {
			msgChan <- "err: powered off"
			continue
		}
		
		if busy && argv[0] != "status" {
			msgChan <- "err: busy -- " + stat()
			continue
		}

		if s.supply > 1000 && argv[0] != "supply" {
			msgChan <- "err: overfull"
			continue
		}

		// Commands master switch
		switch argv[0] {
		
		// Lights
		case "lights":
			switch len(argv) {
			case 1:
				msgChan <- string(s.Lights())
			case 2:
				if argv[1] == "on" {
					s.lights = true
				} else {
					s.lights = false
				}
				ok()
			default:
				invalid()
			}
			
		// Flag
		case "flag":
			switch len(argv) {
			case 1:
				msgChan <- s.flag
			case 2:
				s.flag = argv[1]
				ok()
			default:
				invalid()
			}

		// Contents
		case "contents":
			msgChan <- s.cont
		
		// Power
		case "power":
			switch len(argv) {
			case 1:
				msgChan <- string(s.Power())
			case 2:
				if argv[1] == "on" {
					s.power = true
				} else {
					s.power = false
				}
				ok()
			default:
				invalid()
			}
		
		// Supply
		case "supply":
			switch len(argv) {
			case 1:
				msgChan <- sc.Itoa(s.supply)
			case 3:
				n, err := sc.Atoi(argv[2])

				if err != nil {
					invalid()
				}

				if argv[1] == "load" {
					// TODO ­ more max/min logic
					s.supply += n
					ok()
					go spin(2, "loading", "idle")
				} else {
					// lower
					s.supply -= n
					ok()
					go spin(2, "unloading", "idle")
				}
			default:
				invalid()
			}
		
		// Heat
		case "heat":
			switch len(argv) {
			case 1:
				msgChan <- sc.Itoa(s.temp)
			case 3:
				n, err := sc.Atoi(argv[2])

				if err != nil {
					invalid()
				}

				if argv[1] == "raise" {
					// TODO ­ more max/min logic
					s.temp += n
				} else {
					// lower
					s.temp -= n
				}
				ok()
			default:
				invalid()
			}
		
		// Humidity
		case "humidity":
			switch len(argv) {
			case 1:
				msgChan <- sc.Itoa(s.humid)
			case 3:
				n, err := sc.Atoi(argv[2])

				if err != nil {
					invalid()
				}

				if argv[1] == "raise" {
					// TODO ­ more max/min logic
					s.humid += n
				} else {
					// lower
					s.humid -= n
				}
				ok()
			default:
				invalid()
			}
		
		// Status
		case "status":
			msgChan <- stat()
		
		// Manual disconnect commands, for convenience
		case "quit":
			fallthrough
		case "exit":
			close(msgChan)

		// Command not found
		default:
			msgChan <- "err: unknown command"
		}
	}
	
	if _, more := <- msgChan; more {
		close(msgChan)
	}
	
	log.Println("Silo ended")
}
