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

func connect(_ context.Context) bool {
	log.Infof("enable BLE")

	err := adapter.Enable()
	if err != nil {
		log.Errorf(err.Error())
		time.Sleep(time.Second * 3)

		return false
	}

	ch := make(chan *bluetooth.ScanResult, 1)
	log.Infof("scanning...")
	go func() {
		err := adapter.Scan(func(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
			if app.Canceled {

				err = adapter.StopScan()
				if err != nil {
					log.Errorf(err.Error())
				}
				ch <- nil
			}

			log.Infof("found device:%s %d %s", result.Address.String(), result.RSSI, result.LocalName())
			if result.Address.String() == devAdress.String() {
				err = adapter.StopScan()
				if err != nil {
					log.Errorf(err.Error())
				}
				ch <- &result
			}
		})
		if err != nil {
			log.Errorf(err.Error())
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
				log.Errorf(fmt.Sprintf("%s not found", devAdress.String()))
				time.Sleep(time.Second * 3)
			} else {
				log.Errorf(err.Error())

				return false
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
			return false
		}

		log.Infof("discovering services/characteristics")
		services, err = device.DiscoverServices([]bluetooth.UUID{srvUid})
		if err != nil {
			errCount++
			log.Errorf("%d %v", errCount, err.Error())

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
		log.Errorf("could not find services")
		disconnect()
		time.Sleep(time.Second * 3)

		return false
	}
	service = services[0]

	log.Infof("found servicec: %s", service.UUID().String())

	rx, err := service.DiscoverCharacteristics([]bluetooth.UUID{rxUid})
	if err != nil {
		log.Error(err)
		disconnect()
		time.Sleep(time.Second * 3)

		return false
	}

	if len(rx) == 0 {
		log.Errorf("could not get rx chan")
		disconnect()
		time.Sleep(time.Second * 3)

		return false
	}

	tx, err := service.DiscoverCharacteristics([]bluetooth.UUID{txUid})
	if err != nil {
		log.Errorf(err.Error())
		disconnect()
		time.Sleep(time.Second * 3)

		return false
	}
	if len(tx) == 0 {
		log.Errorf("could not tx characteristic")
		disconnect()
		time.Sleep(time.Second * 3)

		return false
	}

	txChars = tx[0]
	rxChars = rx[0]

	err = rxChars.EnableNotifications(notify)
	if err != nil {
		log.Errorf(err.Error())
		disconnect()

		return false
	}

	return true
}

func notify(buf []byte) {
	if buf[0] == startBit {
		buff = buf
	} else if buf[len(buf)-1] == stopBit {
		buff = append(buff, buf...)
		go read(buff)
		buff = nil
	} else {
		buff = append(buff, buf...)
	}
}
