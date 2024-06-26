package main

import (
	"bleTest/app"
	"bleTest/influx"
	"bleTest/mods"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/godbus/dbus/v5"
	"reflect"
	"time"
)

var (
	lastSendTime time.Time
	isWrited     = false
	ticker       = 0
)

func timeoutCheck() {
	Log.Debugf("timeout check started")

	for {
		if isWrited {
			if time.Since(lastSendTime).Seconds() >= 15 {
				isWrited = false
				Log.Debugf("!!write timeout!!")

				disconnect()
			}
		}

		time.Sleep(time.Second * 5)
	}
}

func writerChan() {
	errCount := 0
	Log.Debugf("start write cycle")

	msg := ReadMessage
	for {
		if app.Canceled {
			break
		}
		Log.Debugf("==================================================================================================================================")
		if ticker%2 == 0 {
			msg = ReadMessage
		} else {
			msg = ReadCellMessage
		}

		resp, err := txChars.WriteWithoutResponse(msg)
		Log.Debugf("Writed: %s %d", hex.EncodeToString(msg), len(msg))

		isWrited = true
		lastSendTime = time.Now()

		msgWG.Add(1)

		if resp == 0 && err != nil {
			var customErr dbus.Error
			if errors.As(err, &customErr) {
				Log.Debugf("error is *dbus.Error")

				if customErr.Error() == NotConnectedError.Error() {
					Log.Errorf(fmt.Errorf("not connected error").Error())
				}

				Log.Errorf(fmt.Errorf("custom error: %v", customErr.Error()).Error())

				disconnect()
				break
			}

			Log.Errorf("unknown error %s :%v", reflect.TypeOf(err).String(), err.Error())
			errCount++
			if errCount > 4 {
				break
			}

			continue
		} else {
			errCount = 0
		}

		Log.Debugf("wait chan...")
		msgWG.Wait()
		isWrited = false
		ticker++

		time.Sleep(3 * time.Second)
	}
}

func recCallb(buf []byte) {
	if buf[0] == StartBit {
		buff = buf
		//log.Debugf("start WTF: %s %d", hex.EncodeToString(buf), len(buf))
	} else if buf[len(buf)-1] == StopBit {
		//log.Debugf("body WTF: %s %d", hex.EncodeToString(buf), len(buf))
		buff = append(buff, buf...)
		Log.Debugf("Received: %s %d", hex.EncodeToString(buff), len(buff))

		go parseData(buff)
		buff = nil
		Log.Debugf("release chan")
		Log.Debugf("==================================================================================================================================")
		msgWG.Done()
	} else {
		buff = append(buff, buf...)
		//log.Debugf("end WTF: %s %d", hex.EncodeToString(buf), len(buf))
	}
}

func parseData(data []byte) {
	if mods.IsValid(data) {
		if data[1] == 0x03 {

			mos := mods.GetMOS(data[24])
			bmsData.Volts = mods.ToFloat([]byte{data[4], data[5]}) / 100
			bmsData.Current = mods.ToFloat([]byte{data[6], data[7]}) / 100
			bmsData.RemainingCapacity = mods.ToFloat([]byte{data[8], data[9]}) / 100
			bmsData.NominalCapcity = mods.ToFloat([]byte{data[10], data[11]}) / 100
			bmsData.Cycles = mods.ToFloat([]byte{data[12], data[13]})
			bmsData.Version = mods.ToVersion(data[22])
			bmsData.RemainingPercent = mods.ToPercents(data[23])
			bmsData.Series = mods.ToInt(data[25])
			//Temp:              toFloat(),
			bmsData.MosChargingEnabled = mos.Charging
			bmsData.MosDischargingEnabled = mos.Discharging

			temp := make([]float32, int(data[26]))
			for i := 0; i < int(data[26]); i++ {
				temperature := (float32(data[27+i*2])*256 + float32(data[28+i*2]) - 2731) / 10
				temp = append(temp, temperature)
			}
			bmsData.Temp = temp

			influx.PushData(bmsData)
		}

		if data[1] == 0x04 {
			// Calculate the number of cells
			bmsNumberOfCells := int(data[3]) / 2

			bmsData.Cells = make([]float32, bmsNumberOfCells)
			// Iterate over each cell
			for i := 0; i < bmsNumberOfCells; i++ {
				index := 4 + 2*i
				millivolts := int(data[index])*256 + int(data[index+1])
				volts := float32(millivolts) / 1000
				bmsData.Cells[0] = volts
				//fmt.Printf("Cell %d: %1.3fV\n", i+1, volts)
			}

			influx.PushCells(bmsData)
		}

	}
}
