package main

import (
	"time"
	"tinygo.org/x/bluetooth"
)

var (
	scanCh   = make(chan *bluetooth.ScanResult, 1)
	Canceled bool
	errCount = 0
	services []bluetooth.DeviceService
)

func updateTimeout() {
	isWrited = true
	lastSendTime = time.Now()
}

func timeoutCompleted() {
	isWrited = false
	lastSendTime = time.Now()
}

func scanCallb(adapter *bluetooth.Adapter, result bluetooth.ScanResult) {
	if Canceled {
		err := adapter.StopScan()
		if err != nil {
			Log.Errorf(err.Error())
		}
		scanCh <- nil
	}

	if result.Address.String() == devAdress.String() {
		Log.Infof("found device:%s %d %s", result.Address.String(), result.RSSI, result.LocalName())
		err := adapter.StopScan()
		if err != nil {
			Log.Errorf(err.Error())
		}
		scanCh <- &result
	}
}

func findBmsDevice() bool {
	Log.Infof("enable BLE")

	err := adapter.Enable()
	if err != nil {
		disconnect(err)
	}

	Log.Infof("looking for: %s ....", devAdress.String())
	go func() {
		err = adapter.Scan(scanCallb)
		if err != nil {
			disconnect(err)
		}
	}()

	var device bluetooth.Device
	select {
	case scanResult := <-scanCh:
		Log.Debugf("try connect")
		updateTimeout()
		device, err = adapter.Connect(scanResult.Address, bluetooth.ConnectionParams{
			ConnectionTimeout: bluetooth.NewDuration(time.Second * 30),
			MinInterval:       bluetooth.NewDuration(50 * time.Millisecond),
			MaxInterval:       bluetooth.NewDuration(100 * time.Millisecond),
			Timeout:           bluetooth.NewDuration(4 * time.Second),
		})
		if err != nil {
			Log.Errorf(err.Error())

			disconnect(err)
		}

		println("connected to ", scanResult.Address.String())
	}
	updateTimeout()

	Log.Debugf("connected: %s", devAdress.String())

	srvUid, err := bluetooth.ParseUUID(serviceUUIDString)
	txUid, err := bluetooth.ParseUUID(txUUIDString)
	rxUid, err := bluetooth.ParseUUID(rxUUIDString)

	for {
		if Canceled {
			return false
		}

		Log.Infof("discovering services/characteristics")
		services, err = device.DiscoverServices([]bluetooth.UUID{srvUid})
		if err != nil {
			errCount++
			Log.Errorf("%d %v", errCount, err.Error())

			if errCount > 3 {
				disconnect(err)

				return false
			}
		} else {
			break
		}
	}

	if len(services) == 0 {
		Log.Errorf("could not find services")
		disconnect()

		return false
	}
	service = &services[0]

	Log.Infof("found servicec: %s", service.UUID().String())

	updateTimeout()

	rx, err := service.DiscoverCharacteristics([]bluetooth.UUID{rxUid})
	if err != nil {
		disconnect(err)

		return false
	}

	updateTimeout()

	if len(rx) == 0 {
		Log.Errorf("could not get rx chan")
		disconnect()

		return false
	}

	tx, err := service.DiscoverCharacteristics([]bluetooth.UUID{txUid})
	if err != nil {
		disconnect(err)

		return false
	}

	updateTimeout()

	if len(tx) == 0 {
		Log.Errorf("could not tx characteristic")
		disconnect()

		return false
	}

	txChars = &tx[0]
	rxChars = &rx[0]

	updateTimeout()

	err = rxChars.EnableNotifications(recCallb)
	if err != nil {
		disconnect(err)

		return false
	}
	Log.Debugf("nofigications enabled")

	timeoutCompleted()

	return true
}
