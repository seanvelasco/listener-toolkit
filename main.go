package main

import (
	"listener-toolkit/database"
	"listener-toolkit/listeners"
)

func main() {

	// create a pointer to an empty struct
	go database.InitializeDatabase()

	go listeners.InitializeTCP("0.0.0.0", "8888")
	go listeners.InitializeUDP("0.0.0.0", "8888")
	go listeners.InitializeHTTP("8080")
	go listeners.InitializeSerial("/dev/cu.usbserial-130", 115200)

	channel := make(chan bool)
	<-channel
}
