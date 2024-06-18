package main

import (
	"bleTest/app"
	"bleTest/logger"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/akamensky/argparse"
	"os"
	"os/exec"
	"runtime"
	"time"
	"tinygo.org/x/bluetooth"
)

var (
	adapter                = bluetooth.DefaultAdapter
	serviceUUIDString      = "0000ff00-0000-1000-8000-00805f9b34fb"
	rxUUIDString           = "0000ff01-0000-1000-8000-00805f9b34fb"
	txUUIDString           = "0000ff02-0000-1000-8000-00805f9b34fb"
	startBit          byte = 0xDD
	stopBit           byte = 0x77
	buff                   = make([]byte, 0)
	rxChars           bluetooth.DeviceCharacteristic
	txChars           bluetooth.DeviceCharacteristic
	devAdress         bluetooth.Address
	service           bluetooth.DeviceService
	log               *logger.Logger
	ctx               context.Context
)

func main() {
	log = logger.New()
	ctx = app.SigTermIntCtx()

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

	for {
		if connect(ctx) && app.Canceled == false {
			writerChan()
		}

		if app.Canceled {
			return
		} else {
			time.Sleep(3 * time.Second)
		}
	}
	log.Debugf("2")
}

func writerChan() {
	var dd = []byte{0xDD, 0xA5, 0x03, 0x00, 0xFF, 0xFD, 0x77}

	errCount := 0

	for {
		if app.Canceled {
			break
		}

		resp, err := txChars.WriteWithoutResponse(dd)
		if resp == 0 {
			log.Errorf(err.Error())
			errCount++
		} else {
			errCount = 0
		}

		if errCount > 10 {
			break
		}

		time.Sleep(3 * time.Second)
	}
	log.Debugf("1")
}

func read(data []byte) {
	log.Debugf("data: %s", hex.EncodeToString(data))
	if data[1] == 0x03 && len(data) >= 26 {

		mos := getMOS(data[24])
		bmsData := JbdData{
			Volts:             toFloat([]byte{data[4], data[5]}) / 100,
			Current:           toFloat([]byte{data[6], data[7]}) / 100,
			RemainingCapacity: toFloat([]byte{data[8], data[9]}) / 100,
			NominalCapcity:    toFloat([]byte{data[10], data[11]}) / 100,
			Cycles:            toFloat([]byte{data[12], data[13]}),
			Version:           toVersion(data[22]),
			RemainingPercent:  toPercents(data[23]),
			Series:            toInt(data[25]),
			//Temp:              toFloat(),
			MosChargingEnabled:    mos.Charging,
			MosDischargingEnabled: mos.Discharging,
		}

		for i := 0; i < int(data[26]); i++ {
			temperature := (float32(data[27+i*2])*256 + float32(data[28+i*2]) - 2731) / 10
			bmsData.Temp = append(bmsData.Temp, temperature)
		}
		//clearConsole()
		//log.Debugf(bmsData.String())
		pushTo(&bmsData)
	}
}

func clearConsole() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	case "linux", "darwin": // для macOS и Linux
		cmd = exec.Command("clear")
	default:
		cmd = exec.Command("clear") // по умолчанию для других Unix-подобных систем
	}

	cmd.Stdout = os.Stdout
	cmd.Run()
}
