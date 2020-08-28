package main

import "fmt"

func warn(format string, values ...interface{}) {
	fmt.Printf("WARN: "+format+"\n", values...)
}
