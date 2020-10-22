package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/tarm/serial"
)

func main() {
	flag.Parse()
	//if verbose {
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Printf("%s=%v\n", f.Name, f.Value)
	})

	settings, err := MakeSettings()
	if err != nil {
		fmt.Printf("while making settings: %v\n", err)
		return
	}

	fmt.Printf("settings: %v\n", *settings)
	config := &serial.Config{
		Name: settings.serialPort,
		Baud: settings.serialBaud,
	}

	serialPort, err := serial.OpenPort(config)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer serialPort.Close()

	serialOut := make(chan string)
	serialIn := make(chan string)
	pipeline := make(chan *Measurement)
	monitor := make(chan int)

	fmt.Println("read the input")
	go ScanLines(serialPort, serialOut, serialIn)
	fmt.Println("measure")
	go Measure(serialOut, pipeline)
	fmt.Println("capture")
	go capture(pipeline, monitor)

	var (
		command string
		ever    bool = true
	)

	printMenu()
	for ever {
		fmt.Scanln(&command)
		switch command {
		case "x":
			ever = false
		case "mr":
			monitor <- MonitorRaw
		case "mx":
			monitor <- MonitorMatrices
		case "o", "c":
			serialIn <- command
		}
		printMenu()
		time.Sleep(time.Millisecond * 10)
	}
}

// monitor status
const (
	MonitorRaw int = iota
	MonitorMatrices
)

func printMenu() {
	fmt.Println("Enter a Command:")
	fmt.Println("  mr - monitor on/off raw data")
	fmt.Println("  mx - monitor on/off matrices")
	fmt.Println("   o - calculate offsets")
	fmt.Println("   c - calculate variances")
	fmt.Println("   x - exit")
}
