package main

import "fmt"

func warnf(format string, values ...interface{}) {
	fmt.Printf("WARN: "+format+"\n", values...)
}
