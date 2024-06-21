package main

import (
	"bleTest/app"
	"bleTest/influx"
	"bleTest/logger"
	"bleTest/mods"
	"errors"
	"fmt"
	"github.com/akamensky/argparse"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
	"tinygo.org/x/bluetooth"
)

var (
	adapter           = *bluetooth.DefaultAdapter
	serviceUUIDString = "0000ff00-0000-1000-8000-00805f9b34fb"
	rxUUIDString      = "0000ff01-0000-1000-8000-00805f9b34fb"
	txUUIDString      = "0000ff02-0000-1000-8000-00805f9b34fb"
	buff              = make([]byte, 50)
	rxChars           *bluetooth.DeviceCharacteristic
	txChars           *bluetooth.DeviceCharacteristic
	devAdress         *bluetooth.Address
	service           *bluetooth.DeviceService
	log               *logger.Logger
	NotConnectedError = errors.New("Not connected")
	AsyncStatus3Error = errors.New("async operation failed with status 3")
	ReadMessage       = []byte{0xDD, 0xA5, 0x03, 0x00, 0xFF, 0xFD, 0x77}
	bmsData           = &mods.JbdData{}
	msgWG             = new(sync.WaitGroup)
)

const StartBit byte = 0xDD
const StopBit byte = 0x77

func handlePanic() {
	if r := recover(); r != nil {
		log.Debugf("Recovered from panic: %v", r)
		// Perform any cleanup or logging here
	}
}

func main() {
	debug.SetGCPercent(10)
	done := make(chan bool, 1)
	defer handlePanic()

	log = logger.New()
	//ctx = app.SigTermIntCtx()

	parser := argparse.NewParser("print", "Prints provided string to stdout")
	s := parser.String("m", "mac", &argparse.Options{Required: false, Help: "required when win or linux"})
	u := parser.String("u", "uuid", &argparse.Options{Required: false, Help: "required when mac"})

	if *s == "" {
		*s = "A5:C2:37:06:1B:C9"
	}

	switch runtime.GOOS {
	case "windows", "linux", "baremetal":
		devAdress = getAdress(*s)
	case "darwin":
		str := "59d9d8cf-7dc9-2f43-ab65-dc2907a5fc4d"
		u = &str

		devAdress = getAdress(*u)
	default:
		fmt.Printf("Current platform is %s\n", runtime.GOOS)
	}
	influx.Init(log)

	go starty()

	<-done

	log.Debugf("Exiting application.")
}

func starty() {
	for {
		if connect() && app.Canceled == false {
			writerChan()
		}

		if app.Canceled {
			break
		} else {
			time.Sleep(3 * time.Second)
		}
	}
}

func disconnect() {
	err := device.Disconnect()
	if err != nil {
		log.Errorf("Error disconnecting device: %v", err)
	}
	err = rxChars.EnableNotifications(nil)
	if err != nil {
		log.Errorf("Error enabling notifications: %v", err)
	}

}
