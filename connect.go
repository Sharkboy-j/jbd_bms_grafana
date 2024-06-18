package main

import (
	"bleTest/app"
	"context"
	"fmt"
	"time"
	"tinygo.org/x/bluetooth"
)

var (
	conErr = fmt.Errorf("device with the given address was not found")
	device bluetooth.Device
)

func connect(ctx context.Context) {
	log.Infof("enable BLE")
	err := adapter.Enable()
	if err != nil {
		log.Errorf(err.Error())
		time.Sleep(time.Second * 3)

		return
	}

	ch := make(chan *bluetooth.ScanResult, 1)
	log.Infof("scanning...")
	go adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
		if app.Canceled {
			adapter.StopScan()
			ch <- nil
		}

		log.Infof("found device:%s %d %s", result.Address.String(), result.RSSI, result.LocalName())
		if result.Address.String() == devAdress.String() {
			adapter.StopScan()
			ch <- &result
		}
	})

	result := <-ch

	for {
		if app.Canceled {
			return
		}

		device, err = adapter.Connect(result.Address, bluetooth.ConnectionParams{
			ConnectionTimeout: bluetooth.NewDuration(time.Second * 30),
			MinInterval:       bluetooth.NewDuration(495 * time.Millisecond),
			MaxInterval:       bluetooth.NewDuration(510 * time.Millisecond),
			Timeout:           bluetooth.NewDuration(10 * time.Second),
		})

		if err != nil {
			if err.Error() == conErr.Error() {
				log.Errorf(fmt.Sprintf("%s not found", devAdress.String()))
				time.Sleep(time.Second * 3)
			} else {
				log.Errorf(err.Error())

				return
			}
		} else {
			break
		}
	}

	log.Infof("connected: %s", devAdress.String())

	srvUid, err := bluetooth.ParseUUID(serviceUUIDString)
	txUid, err := bluetooth.ParseUUID(txUUIDString)
	rxUid, err := bluetooth.ParseUUID(rxUUIDString)

	var errCount = 0

	var services []bluetooth.DeviceService
	for {
		if app.Canceled {
			return
		}

		log.Infof("discovering services/characteristics")
		services, err = device.DiscoverServices([]bluetooth.UUID{srvUid})
		if err != nil {
			log.Errorf(err.Error())
			errCount++

			if errCount > 10 {
				return
			}
			time.Sleep(time.Second * 1)
		} else {
			break
		}
	}

	if len(services) == 0 {
		log.Errorf("could not find services")
		device.Disconnect()
		time.Sleep(time.Second * 3)

		return
	}
	service = services[0]

	log.Infof("found servicec: %s", service.UUID().String())

	rx, err := service.DiscoverCharacteristics([]bluetooth.UUID{rxUid})
	if err != nil {
		log.Errorf(err.Error())
		device.Disconnect()
		time.Sleep(time.Second * 3)

		return
	}

	if len(rx) == 0 {
		log.Errorf("could not get rx chan")
		device.Disconnect()
		time.Sleep(time.Second * 3)

		return
	}

	tx, err := service.DiscoverCharacteristics([]bluetooth.UUID{txUid})
	if err != nil {
		log.Errorf(err.Error())
		device.Disconnect()
		time.Sleep(time.Second * 3)

		return
	}
	if len(tx) == 0 {
		log.Errorf("could not tx characteristic")
		device.Disconnect()
		time.Sleep(time.Second * 3)

		return
	}

	txChars = tx[0]
	rxChars = rx[0]

	err = rxChars.EnableNotifications(notify)
	if err != nil {
		log.Errorf(err.Error())
		device.Disconnect()

		return
	}
}

func notify(buf []byte) {
	if buf[0] == startBit {
		buff = buf
	} else if buf[len(buf)-1] == stopBit {
		buff = append(buff, buf...)
		read(buff)
		buff = []byte{}
	} else {
		buff = append(buff, buf...)
	}
}
