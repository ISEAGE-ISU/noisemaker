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

