package main

import (
	"fmt"
	"os"
	"time"

	"grd.qbus/qbus"
)

func main() {
	sv, e := qbus.QBusInit()
	c, _ := qbus.NewQBusClient("com.qbusclient.test")
	c2, _ := qbus.NewQBusClient("com.qbusclient2.test")
	fmt.Println(e)

	go sv.Open()

	time.Sleep(time.Second)
	go c.Open()
	time.Sleep(time.Second)
	go c2.Open()
	time.Sleep(time.Second)
	os.Exit(0)
}
