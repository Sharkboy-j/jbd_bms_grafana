package main

import (
	"bleTest/app"
	"bleTest/logger"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/akamensky/argparse"
	"github.com/godbus/dbus/v5"
	"runtime"
	"runtime/debug"
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
	bmsData           = &JbdData{}
	lastInd           = 0
	MSGcH             = make(chan bool, 1)
)

const startBit byte = 0xDD
const stopBit byte = 0x77

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
	initData()

	go starty()

	//for {
	//	log.Debugf(time.Now().String())
	//	time.Sleep(time.Second * 1)
	//}

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

func writerChan() {
	errCount := 0

	for {
		if app.Canceled {
			break
		}

		resp, err := txChars.WriteWithoutResponse(ReadMessage)
		if resp == 0 && err != nil {
			var customErr *dbus.Error
			if errors.As(err, &customErr) {
				if customErr.Error() == NotConnectedError.Error() {
					log.Errorf(fmt.Errorf("not connected error").Error())
					disconnect()

					break
				} else {
					log.Errorf(fmt.Errorf("custom error: %v", customErr.Error()).Error())
				}
			} else {
				if errors.Is(err, AsyncStatus3Error) {
					disconnect()

					break
				}
			}

			log.Errorf(err.Error())
			errCount++
		} else {
			errCount = 0
		}

		if errCount > 4 {
			break
		}

		log.Debugf("wait chan")
		<-MSGcH
		time.Sleep(3 * time.Second)
	}
}

func parseData() {
	if isValid() {
		log.Debugf("data: %s %d", hex.EncodeToString(buff), len(buff))
		if buff[1] == 0x03 && len(buff) >= 26 {

			mos := getMOS(buff[24])
			bmsData.Volts = toFloat([]byte{buff[4], buff[5]}) / 100
			bmsData.Current = toFloat([]byte{buff[6], buff[7]}) / 100
			bmsData.RemainingCapacity = toFloat([]byte{buff[8], buff[9]}) / 100
			bmsData.NominalCapcity = toFloat([]byte{buff[10], buff[11]}) / 100
			bmsData.Cycles = toFloat([]byte{buff[12], buff[13]})
			bmsData.Version = toVersion(buff[22])
			bmsData.RemainingPercent = toPercents(buff[23])
			bmsData.Series = toInt(buff[25])
			//Temp:              toFloat(),
			bmsData.MosChargingEnabled = mos.Charging
			bmsData.MosDischargingEnabled = mos.Discharging

			temp := make([]float32, int(buff[26]))
			for i := 0; i < int(buff[26]); i++ {
				temperature := (float32(buff[27+i*2])*256 + float32(buff[28+i*2]) - 2731) / 10
				temp = append(temp, temperature)
			}
			bmsData.Temp = temp

			//clearConsole()
			//log.Debugf(bmsData.String())
			pushTo(bmsData)
		}
	}
}

func isValid() bool {
	return buff[0] == startBit && buff[len(buff)-1] == stopBit
}
