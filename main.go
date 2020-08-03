package main

import (
	"os"
	"src/src"
)

func main() {

	// err := src.XMLToTickData("20200203")
	start := ""
	if len(os.Args) > 1 {
		start = os.Args[1]
	}
	src.Run(start)
}
