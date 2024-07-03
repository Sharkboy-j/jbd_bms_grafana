package main

import (
	"bleTest/bluetoothHelper"
	"bleTest/influx"
	"bleTest/logger"
	"bleTest/mods"
	"errors"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"sync"
	"time"
	"tinygo.org/x/bluetooth"
)

//goland:noinspection GoErrorStringFormat
var (
	adapter           = *bluetooth.DefaultAdapter
	buff              = make([]byte, 50)
	rxChars           *bluetooth.DeviceCharacteristic
	txChars           *bluetooth.DeviceCharacteristic
	devAdress         *bluetooth.Address
	service           *bluetooth.DeviceService
	Log               *logger.Logger
	NotConnectedError = errors.New("Not connected")
	ReadMessage       = []byte{0xDD, 0xA5, 0x03, 0x00, 0xFF, 0xFD, 0x77}
	ReadCellMessage   = []byte{0xdd, 0xa5, 0x4, 0x0, 0xff, 0xfc, 0x77}
	bmsData           = &mods.JbdData{}
	msgWG             = new(sync.WaitGroup)
)

const (
	StartBit          byte = 0xDD
	StopBit           byte = 0x77
	macAddrStr             = "A5:C2:37:06:1B:C9"
	uidAddrStr             = "59d9d8cf-7dc9-2f43-ab65-dc2907a5fc4d"
	serviceUUIDString      = "0000ff00-0000-1000-8000-00805f9b34fb"
	rxUUIDString           = "0000ff01-0000-1000-8000-00805f9b34fb"
	txUUIDString           = "0000ff02-0000-1000-8000-00805f9b34fb"
)

func main() {
	debug.SetGCPercent(10)
	done := make(chan bool, 1)
	Log = logger.New()
	Log.Infof("Entry point")

	s := os.Getenv("BMS_MAC")
	_ = os.Getenv("TIMEOUT")
	u := os.Getenv("BMS_UUID")

	switch runtime.GOOS {
	case "windows", "linux", "baremetal":
		if s == "" {
			s = macAddrStr
		}

		devAdress = bluetoothHelper.GetAdress(Log, s)
	case "darwin":
		if u == "" {
			u = uidAddrStr
		}

		devAdress = bluetoothHelper.GetAdress(Log, u)
	default:
		fmt.Printf("Current platform is %s\n", runtime.GOOS)
	}
	influx.Init(Log)
	go starty()

	<-done
	Log.Infof("Exiting application.")
}

func timeoutCheck() {
	Log.Debugf("timeout check started")

	for {
		if isWrited {
			if time.Since(lastSendTime).Seconds() >= 5 {
				timeoutCompleted()

				Log.Warnf("!!timeout!!")

				disconnect()
			}
		}

		time.Sleep(time.Second * 5)
	}
}

func starty() {
	go timeoutCheck()

	if findBmsDevice() && Canceled == false {
		writerChan()
	}
}

func disconnect(err ...error) {
	Log.Errorf("restart due to shit:")
	if len(err) > 0 {
		panic(err[0].Error())
	} else {
		panic(0)
	}
}
