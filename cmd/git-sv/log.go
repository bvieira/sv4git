package main

import (
	"fmt"
	"os"
)

func warnf(format string, values ...interface{}) {
	fmt.Fprintf(os.Stderr, "WARN: "+format+"\n", values...)
}
