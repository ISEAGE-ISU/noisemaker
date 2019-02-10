package main

import (
	"fmt"
	"flag"
)

// Generate diagnostics, health, geo-location info to send to accounting system. 
func main() {
	var confName, hostsName string

	flag.StringVar(&confName, "c", "config.ndb", "Config ndb file name")
	flag.StringVar(&hostsName, "n", "hosts.ndb", "Hosts ndb file name")
	flag.BoolVar(&conf.chatty, "D", true, "Debug chattiness")

	flag.Parse()
	
	// Init conf
	loadNdbs(confName, hostsName)

	fmt.Println("Init…")

}

