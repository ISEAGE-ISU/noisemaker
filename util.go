package main

import (
	"fmt"
	"os"
)

// Check if err != nil, if so, sysfatal with the relevant message
func efatal(err error, msg string) {
	if err != nil {
		sysfatal("Error: %s %s", msg, err.Error())
	}
}

// Emulate sysfatal(2) from plan9 (exit with a message)
func sysfatal(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", os.Args[0], fmt.Sprintf(format, args...))
	os.Exit(2)
}

// Stringify a given object
func stringify(x interface{}) string {
	return fmt.Sprintf("%v", x)
}

// Bool to string
func b2s(b bool) string {
	if b {
		return "on"
	}

	return "off"
}

// String to bool
func s2b(s string) bool {
	switch s {
	case "on":
		return true
	default:
		return false
	}
}

// Check if s contains a t
func in(s []string, t string) bool {
	for _, v := range s {
		if v == t {
			return true
		}
	}
	
	return false
}
