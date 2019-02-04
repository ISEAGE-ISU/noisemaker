package main

import (
	"time"
)

type Config struct {
	chatty		bool				// Control chattiness of debug output
	waitmin		time.Duration		// Wait time (minimum) between transmissions to a host
	hosts		map[string]string	// Hosts table of short name:address
}

// Global config
var conf Config
