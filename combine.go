package main

import (
	"strings"
	"net"
sc	"strconv"
)

// Combine type
type Combine struct {
	lights	bool		// whether combine lights are on or not
	power	bool		// power on/off
	status	string	// current action
	humid	int		// current humidity %
	temp	int		// current temp °C
	supply	int		// current supply levels bushels
	cont		string	// current contents
	fuel		int		// fuel %
	oil		int		// oil %	
	tires		int		// tires %
}

// Printable lights format
func (t *Combine) Lights() string {
	return b2s(t.lights)
}

// Printable lights format
func (t *Combine) Power() string {
	return b2s(t.power)
}

// Constructor Combine
func NewCombine() (t Combine) {
	// Init combine
	status	= "idle"
	t.humid	= 30
	t.temp	= 20
	t.supply	= 0
	t.cont	= "corn"
	t.fuel 	= 0
	t.oil		= 0
	t.tires	= 100

	return
}

func (t *Combine) DoCmd(conn net.Conn, connected *bool, write func(string), read func([]byte), invalid func()) {
	ok := func() { write("ok.") }

	for *connected {
		buf := make([]byte, width)
		
		conn.Write([]byte("> "))

		read(buf)

		argv := strings.Fields(string(buf))
		
		if len(argv) > 1 {
			// '\n' counts as a field split, truncate
			argv = argv[:len(argv)-1]
		}

		if len(argv) < 1 {
			write("err: no command specified")
		}
		
		if argv[0] == "power" || argv[0] == "fuel" || argv[0] == "oil" {
			goto cmd
		}

		if !t.power {
			write("err: powered off")
			goto nocmd
		}

		if busy && argv[0] != "status" {
			write("err: busy -- " + stat())
			goto nocmd
		}

		if t.fuel <= 0 && argv[0] != "fuel" {
			write("err: no fuel")
			goto nocmd
		}

		if t.oil <= 0 && argv[0] != "oil" {
			write("err: no oil")
			goto nocmd
		}

		if t.supply > 100 && argv[0] != "supply" {
			write("err: overfull")
			goto nocmd
		}

		cmd:

		// Commands master switch
		switch argv[0] {
		
		// Lights
		case "lights":
			switch len(argv) {
			case 1:
				write(string(t.Lights()))
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
			
		// Oil
		case "oil":
			switch len(argv) {
			case 1:
				write(sc.Itoa(t.oil))
			case 2:
				if argv[1] == "add" {
					t.oil += 100
					ok()
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
				write(sc.Itoa(t.fuel))
			case 2:
				if argv[1] == "add" {
					t.fuel += 100
					ok()
					go spin(5, "refueling", "idle")
				} else {
					t.fuel = 0
					ok()
				}
			default:
				invalid()
			}
			
		// Tires
		case "tires":
			switch len(argv) {
			case 1:
				write(sc.Itoa(t.tires))
				ok()
			case 2:
				if argv[1] == "inflate" {
					t.tires = 100
				} else {
					t.tires = 0
				}
				ok()
			default:
				invalid()
			}
		
		// Harvest
		case "harvest":
			switch len(argv) {
				case 1:
					write(t.status)
				case 2:
					if argv[1] == "start" {
						ok()
						go spin(10, "harvesting", "idle")
					} else {
						ok()
						go spin(10, "harvesting", "idle")
					}
				default:
					invalid()
				}

		// Contents
		case "contents":
			write(t.cont)
		
		// Power
		case "power":
			switch len(argv) {
			case 1:
				write(string(t.Power()))
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
				write(sc.Itoa(t.supply))
			case 3:
				n, err := sc.Atoi(argv[2])

				if err != nil {
					invalid()
				}

				if argv[1] == "load" {
					// TODO ­ more max/min logic
					t.supply += n
					ok()
					go spin(10, "loading", "idle")
				} else {
					// lower
					t.supply -= n
					ok()
					go spin(10, "unloading", "idle")
				}
			default:
				invalid()
			}
		
		// Heat
		case "heat":
			switch len(argv) {
			case 1:
				write(sc.Itoa(t.temp))
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
				write(sc.Itoa(t.humid))
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
			write(stat())
		
		// Manual disconnect commands, for convenience
		case "quit":
			fallthrough
		case "exit":
			ok()
			return
			break

		// Command not found
		default:
			write("err: unknown command")
		}
		nocmd:
	}
}
