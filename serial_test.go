package main

import (
	"testing"
	"time"

	"github.com/tarm/serial"
)

const serialFileName = "/dev/ttyUSB0"

func testSerial(t *testing.T) {
	settings, err := MakeSettings()
	config := &serial.Config{Name: settings.serialPort, Baud: settings.serialBaud}
	port, err := serial.OpenPort(config)

	if err != nil {
		t.Fatalf("%v", err)
	}
	defer port.Close()

	out := make(chan string)
	in := make(chan string)
	quitScan := make(chan int)
	quitPrint := make(chan int)
	go ScanLines(port, out, in)
	time.Sleep(time.Second * 10)
	quitScan <- 1
	quitPrint <- 1
	time.Sleep(time.Second * 1)

}
