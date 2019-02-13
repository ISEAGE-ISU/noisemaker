package main

// Silo type
type Silo struct {
	lights	bool	// whether silo lights are on or not
	status	string	// current action
	humid	int		// current humidity mmHg
	temp	int		// current temp °C
	supply	int		// current supply levels
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
