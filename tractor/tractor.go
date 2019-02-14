package main

// Tractor type
type Tractor struct {
	lights	bool		// whether tractor lights are on or not
	power	bool		// power on/off
	status	string	// current action
	humid	int		// current humidity %
	temp	int		// current temp °C
	supply	int		// current supply levels bushels
	cont		string	// current contents
	pin		int		// current auth pin
	auth		bool		// session authenticated y/n
	fuel		int		// fuel %
	oil		int		// oil %	
	tires		int		// tires %
}

// Printable lights format
func (s *Tractor) Lights() string {
	if s.lights {
		return "on"
	}
	
	return "off"
}

// Printable lights format
func (s *Tractor) Power() string {
	if s.power {
		return "on"
	}
	
	return "off"
}
