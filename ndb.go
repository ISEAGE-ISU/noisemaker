package main

import (
	"github.com/mischief/ndb"
	"fmt"
	"os"
sc	"strconv"
	"time"
)

// Convert a RecordSet into a map[string]string
func rs2map(r ndb.RecordSet) (m map[string]string) {
	m = make(map[string]string)

	for _, rec := range r {
		var rs ndb.RecordSet = []ndb.Record{ rec }
		sys := rs.Search("sys")
		host := rs.Search("host")
		m[sys] = host
	}
	
	return
}

// Look up a singleton tuple ­ no children, no duplicates
func loadOne(n *ndb.Ndb, attr, def string) string {
	str := n.Search(attr, "").Search(attr)

	if str == "" {
		fmt.Fprintf(os.Stderr, "Warning: failed to find %s tuple.\n", attr)
		return def
	} 

	return str
}

// Load ndb config and hosts
func loadNdbs(confName, hostsName string) {
	// config
	cndb, err := ndb.Open(confName)
	efatal(err, "config file parsing failed.")
	
	// Load wait time (in seconds)
	waittimes := loadOne(cndb, "waittime", "10")
	wti, err := sc.Atoi(waittimes)
	if err != nil {
		efatal(err, "invalid waittimes value")
	}
	conf.waitmin = time.Duration(wti) * time.Second

	// Hosts
	conf.hosts = make(map[string]string)
	hndb, err := ndb.Open(hostsName)
	efatal(err, "hosts file parsing failed.")

	conf.hosts = rs2map(hndb.Search("sys", ""))
}

