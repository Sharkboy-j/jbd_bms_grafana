package main

import (
	"bleTest/app"
	"fmt"
	"time"
	"tinygo.org/x/bluetooth"
)

var (
	conErr = fmt.Errorf("device with the given address was not found")
	device bluetooth.Device
)

func connect() bool {
	Log.Infof("enable BLE")

	err := adapter.Enable()
	if err != nil {
		Log.Errorf(err.Error())
		time.Sleep(time.Second * 3)

		return false
	}

	ch := make(chan *bluetooth.ScanResult, 1)
	Log.Infof("scanning...")
	go func() {
		err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			if app.Canceled {

				err = adapter.StopScan()
				if err != nil {
					Log.Errorf(err.Error())
				}
				ch <- nil
			}

			Log.Infof("found device:%s %d %s", result.Address.String(), result.RSSI, result.LocalName())
			if result.Address.String() == devAdress.String() {
				err = adapter.StopScan()
				if err != nil {
					Log.Errorf(err.Error())
				}
				ch <- &result
			}
		})
		if err != nil {
			Log.Errorf(err.Error())
		}
	}()

	result := <-ch

	for {
		if app.Canceled {
			return false
		}

		device, err = adapter.Connect(result.Address, bluetooth.ConnectionParams{
			ConnectionTimeout: bluetooth.NewDuration(time.Second * 30),
			MinInterval:       bluetooth.NewDuration(495 * time.Millisecond),
			MaxInterval:       bluetooth.NewDuration(510 * time.Millisecond),
			Timeout:           bluetooth.NewDuration(10 * time.Second),
		})

		if err != nil {
			if err.Error() == conErr.Error() {
				Log.Errorf(fmt.Sprintf("%s not found", devAdress.String()))
				time.Sleep(time.Second * 3)
			} else {
				Log.Errorf(err.Error())

				return false
			}
		} else {
			break
		}
	}

	Log.Infof("connected: %s", devAdress.String())

	srvUid, err := bluetooth.ParseUUID(serviceUUIDString)
	txUid, err := bluetooth.ParseUUID(txUUIDString)
	rxUid, err := bluetooth.ParseUUID(rxUUIDString)

	var errCount = 0

	var services []bluetooth.DeviceService
	for {
		if app.Canceled {
			return false
		}

		Log.Infof("discovering services/characteristics")
		services, err = device.DiscoverServices([]bluetooth.UUID{srvUid})
		if err != nil {
			errCount++
			Log.Errorf("%d %v", errCount, err.Error())

			if errCount > 10 {
				disconnect()

				return false
			}

			time.Sleep(time.Second * 1)
		} else {
			break
		}
	}

	if len(services) == 0 {
		Log.Errorf("could not find services")
		disconnect()
		time.Sleep(time.Second * 3)

		return false
	}
	service = &services[0]

	Log.Infof("found servicec: %s", service.UUID().String())

	rx, err := service.DiscoverCharacteristics([]bluetooth.UUID{rxUid})
	if err != nil {
		Log.Error(err)
		disconnect()
		time.Sleep(time.Second * 3)

		return false
	}

	if len(rx) == 0 {
		Log.Errorf("could not get rx chan")
		disconnect()
		time.Sleep(time.Second * 3)

		return false
	}

	tx, err := service.DiscoverCharacteristics([]bluetooth.UUID{txUid})
	if err != nil {
		Log.Errorf(err.Error())
		disconnect()
		time.Sleep(time.Second * 3)

		return false
	}
	if len(tx) == 0 {
		Log.Errorf("could not tx characteristic")
		disconnect()
		time.Sleep(time.Second * 3)

		return false
	}

	txChars = &tx[0]
	rxChars = &rx[0]

	if isRecCallbEnabled == false {
		err = rxChars.EnableNotifications(recCallb)
		if err != nil {
			Log.Errorf(err.Error())
			disconnect()

			return false
		}
		Log.Debugf("nofigications enabled")
	}

	return true
}
