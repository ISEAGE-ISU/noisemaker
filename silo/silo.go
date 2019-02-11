package main

// Silo type
type Silo struct {
	lights	bool	// whether silo lights are on or not
}

// Printable lights format
func (s *Silo) Lights() string {
	if s.lights {
		return "on"
	}
	
	return "off"
}
