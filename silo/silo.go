package main

// Silo type
type Silo struct {
	lights	bool	// whether silo lights are on or not
	power	bool	// power on/off
	status	string	// current action
	humid	int		// current humidity %
	temp	int		// current temp °C
	supply	int		// current supply levels bushels
	cont	string	// current contents
	pin		int		// current auth pin
	auth	bool	// session authenticated y/n
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
	if s.power {
		return "on"
	}
	
	return "off"
}
