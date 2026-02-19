package main

import (
	"fmt"
	"os"
	"time"

	"grd.buildTools/sbuild"
)

var VERSION = 1
var VERSION_STR = "1.0"

func main() {
	fmt.Println("buildTools for mlim version", VERSION_STR)

	startTime := time.Now()
	e := sbuild.Ship()

	endTime := time.Now()

	fmt.Println("Build done in", endTime.Sub(startTime).Seconds(), "(s)")

	if e != nil {
		os.Exit(-1)
	}
}
