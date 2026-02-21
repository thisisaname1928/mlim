package main

import (
	"fmt"
	"time"

	"grd.qbus/qbus"
)

func main() {
	sv, e := qbus.QBusInit()
	c, _ := qbus.NewQBusClient()
	fmt.Println(e)

	go sv.Open()

	time.Sleep(time.Second * 2)
	c.Open()
	time.Sleep(time.Second * 2)
}
