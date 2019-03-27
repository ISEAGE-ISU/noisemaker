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

// Creates a cfg map to write to a file
func (s *Silo) Cfg() map[string]string {
	return map[string]string{
	"supply": sc.Itoa(s.supply),
	"cont": s.cont,
	"flag": s.flag,
	"pin": sc.Itoa(pin),
	}
}

// Processes a cfg map to load into current state
func (s *Silo) LoadCfg(cfg map[string]string) {
	if len(cfg) < 1 {
		return
	}

	// Need: supply, contents, flag
	s.flag = cfg["flag"]
	s.cont = cfg["cont"]
	i64, _ := sc.ParseInt(cfg["supply"], 10, 32)
	s.supply = int(i64)
	pin, _ = sc.Atoi(cfg["pin"])
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
		
		if argv[0] == "power" || argv[0] == "supply" || argv[0] == "status" {
			goto cmd
		}
		
		if !s.power {
			msgChan <- "err: powered off"
			continue
		}
		
		if busy {
			msgChan <- "err: busy -- " + stat()
			continue
		}

		if s.supply > 1000 {
			msgChan <- "err: overfull"
			continue
		}
		
		cmd:

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
		
		// PIN
		case "pin":
			switch len(argv) {
			case 1:
				msgChan <- string(pin)
			case 2:
				if p, err := sc.Atoi(argv[1]); err != nil {
					invalid()
				} else {
					pin = p
					ok()
					dumpChan <- s.Cfg()
				}
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
				dumpChan <- s.Cfg()
			default:
				invalid()
			}

		// Contents
		case "contents":
			switch len(argv) {
			case 1:
				msgChan <- s.cont
			case 2:
				if in(contTypes, argv[1]) {
					s.cont = argv[1]
					ok()
					dumpChan <- s.Cfg()
				} else {
					invalid()
				}
			default:
				invalid()
			}
		
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
					if s.supply + n > 1000 {
						msgChan <- "err: Supply would go overfull!"
						break
					} else {
						s.supply += n
						ok()
						go spin(2, "loading", "idle")
					}
				} else {
					// lower
					if s.supply - n < 0 {
						msgChan <- "err: Supply may not go negative!"
						break
					} else {
						s.supply -= n
						ok()
						go spin(2, "unloading", "idle")
					}
				}
				
				dumpChan <- s.Cfg()
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
					break
				}

				if argv[1] == "raise" {
					if s.temp + n > 100 {
						msgChan <- "err: Temp would exceed 100°C!"
						break
					} else {
						s.temp += n
					}
				} else {
					// lower
					if s.temp - n < 0 {
						msgChan <- "err: Temp would be negative!"
						break
					} else {
						s.temp -= n
					}
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
					break
				}

				if argv[1] == "raise" {
					if s.humid + n > 100 {
						msgChan <- "err: Humidity would exceed max!"
						break
					} else {
						s.humid += n
					}
				} else {
					// lower
					if s.humid - n < 0 {
						msgChan <- "err: Humidity would be negative!"
						break
					} else {
						s.humid -= n
					}
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
