package main

import (
	"strings"
	"log"
sc	"strconv"
)

// Tractor type
type Tractor struct {
	lights	bool		// whether tractor lights are on or not
	power	bool		// power on/off
	humid	int		// current humidity %
	temp	int		// current temp °C
	supply	int		// current supply levels bushels
	cont		string	// current contents
	fuel		int		// fuel %
	oil		int		// oil %	
	tires		int		// tires %
	flag	string	// flag string
}

// Creates a cfg map to write to a file
func (t *Tractor) Cfg() map[string]string {
	return map[string]string{
	"supply": sc.Itoa(t.supply),
	"cont": t.cont,
	"flag": t.flag,
	"fuel": sc.Itoa(t.fuel),
	"oil": sc.Itoa(t.oil),
	"tires": sc.Itoa(t.tires),
	}
}

// Processes a cfg map to load into current state
func (t *Tractor) LoadCfg(cfg map[string]string) {
	if len(cfg) < 1 {
		return
	}

	// Need: supply, contents, flag, fuel, oil, tires
	t.flag = cfg["flag"]
	t.cont = cfg["cont"]

	i64, _ := sc.ParseInt(cfg["supply"], 10, 32)
	t.supply = int(i64)
	
	i64, _ = sc.ParseInt(cfg["fuel"], 10, 32)
	t.fuel = int(i64)
	
	i64, _ = sc.ParseInt(cfg["oil"], 10, 32)
	t.oil = int(i64)
	
	i64, _ = sc.ParseInt(cfg["tires"], 10, 32)
	t.tires = int(i64)
}

// Printable lights format
func (t *Tractor) Lights() string {
	return b2s(t.lights)
}

// Printable lights format
func (t *Tractor) Power() string {
	return b2s(t.power)
}

// Constructor Tractor
func NewTractor() (t Tractor) {
	// Init tractor
	t.humid	= 30
	t.temp	= 20
	t.supply	= 0
	t.cont	= "corn"
	t.fuel 	= 0
	t.oil		= 0
	t.tires	= 100

	return
}

func (t *Tractor) DoCmd(msgChan chan string) {
	ok := func() { msgChan <- "ok." }
	invalid := func() { msgChan <- "err: invalid arguments" }

	for {
		log.Println("New loop")
	
		buf, more := <- msgChan
		if !more {
			break 
		}

		argv := strings.Fields(string(buf))
		
		if len(argv) > 1 {
			// '\n' counts as a field split, truncate
			argv = argv[:len(argv)-1]
		}

		if len(argv) < 1 {
			msgChan <- "err: no command specified"
			continue
		}
		
		if argv[0] == "power" || argv[0] == "fuel" || argv[0] == "oil" {
			goto cmd
		}

		if !t.power {
			msgChan <- "err: powered off"
			continue
		}

		if busy && argv[0] != "status" {
			msgChan <- "err: busy -- " + stat()
			continue
		}

		if t.fuel <= 0 && argv[0] != "fuel" {
			msgChan <- "err: no fuel"
			continue
		}

		if t.oil <= 0 && argv[0] != "oil" {
			msgChan <- "err: no oil"
			continue
		}

		if t.supply > 100 && argv[0] != "supply" {
			msgChan <- "err: overfull"
			continue
		}

		cmd:
		
		log.Println("start switch")

		// Commands master switch
		switch argv[0] {
		
		// Lights
		case "lights":
			switch len(argv) {
			case 1:
				msgChan <- string(t.Lights())
			case 2:
				if argv[1] == "on" {
					t.lights = true
				} else {
					t.lights = false
				}
				ok()
			default:
				invalid()
			}
		
		// Flag
		case "flag":
			switch len(argv) {
			case 1:
				msgChan <- t.flag
			case 2:
				t.flag = argv[1]
				ok()
				dumpChan <- t.Cfg()
			default:
				invalid()
			}
			
		// Oil
		case "oil":
			switch len(argv) {
			case 1:
				msgChan <- sc.Itoa(t.oil)
			case 2:
				if argv[1] == "add" {
					t.oil += 100
					ok()
					dumpChan <- t.Cfg()
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
				msgChan <- sc.Itoa(t.fuel)
			case 2:
				if argv[1] == "add" {
					t.fuel += 100
					ok()
					go spin(1, "refueling", "idle")
				} else {
					t.fuel = 0
					ok()
				}
				dumpChan <- t.Cfg()
			default:
				invalid()
			}
			
		// Tires
		case "tires":
			switch len(argv) {
			case 1:
				msgChan <- sc.Itoa(t.tires)
				ok()
			case 2:
				if argv[1] == "inflate" {
					t.tires = 100
				} else {
					t.tires = 0
				}
				ok()
				dumpChan <- t.Cfg()
			default:
				invalid()
			}
		
		// Harvest
		case "harvest":
			switch len(argv) {
				case 1:
					msgChan <- stat()
				case 2:
					if argv[1] == "start" {
						ok()
						go spin(2, "harvesting", "idle")
					} else {
						ok()
						go spin(2, "harvesting", "idle")
					}
				default:
					invalid()
				}

		// Contents
		case "contents":
			switch len(argv) {
			case 1:
				msgChan <- t.cont
			case 2:
				t.cont = argv[1]
				ok()
				dumpChan <- t.Cfg()
			default:
				invalid()
			}
		
		// Power
		case "power":
			switch len(argv) {
			case 1:
				msgChan <- string(t.Power())
			case 2:
				if argv[1] == "on" {
					t.power = true
				} else {
					t.power = false
				}
				ok()
			default:
				invalid()
			}
		
		// Supply
		case "supply":
			switch len(argv) {
			case 1:
				msgChan <- sc.Itoa(t.supply)
			case 3:
				n, err := sc.Atoi(argv[2])

				if err != nil {
					invalid()
				}

				if argv[1] == "load" {
					// TODO ­ more max/min logic
					t.supply += n
					ok()
					go spin(2, "loading", "idle")
				} else {
					// lower
					t.supply -= n
					ok()
					go spin(2, "unloading", "idle")
				}
				
				dumpChan <- t.Cfg()
			default:
				invalid()
			}
		
		// Heat
		case "heat":
			switch len(argv) {
			case 1:
				msgChan <- sc.Itoa(t.temp)
			case 3:
				n, err := sc.Atoi(argv[2])

				if err != nil {
					invalid()
				}

				if argv[1] == "raise" {
					// TODO ­ more max/min logic
					t.temp += n
				} else {
					// lower
					t.temp -= n
				}
				ok()
			default:
				invalid()
			}
		
		// Humidity
		case "humidity":
			switch len(argv) {
			case 1:
				msgChan <- sc.Itoa(t.humid)
			case 3:
				n, err := sc.Atoi(argv[2])

				if err != nil {
					invalid()
				}

				if argv[1] == "raise" {
					// TODO ­ more max/min logic
					t.humid += n
				} else {
					// lower
					t.humid -= n
				}
				ok()
			default:
				invalid()
			}
		
		// Status
		case "status":
			log.Println("reached status")
			msgChan <- stat()
			log.Println("left status")
		
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
	
	log.Println("Tractor ended")
}
