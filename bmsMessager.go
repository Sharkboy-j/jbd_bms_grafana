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
)

func timeoutCheck() {
	for {
		if isWrited {
			if time.Since(lastSendTime).Seconds() >= 30 {
				log.Debugf("!!write timeout!!")
				msgWG.Done()
			}
		}

		time.Sleep(time.Second * 5)
	}
}

func writerChan() {
	errCount := 0
	log.Debugf("start write cycle")

	for {
		if app.Canceled {
			break
		}

		resp, err := txChars.WriteWithoutResponse(ReadMessage)
		isWrited = true
		lastSendTime = time.Now()

		log.Debugf("writed")
		msgWG.Add(1)

		if resp == 0 && err != nil {
			var customErr dbus.Error
			if errors.As(err, &customErr) {
				log.Debugf("error is *dbus.Error")
				if customErr.Error() == NotConnectedError.Error() {
					log.Errorf(fmt.Errorf("not connected error").Error())
					disconnect()

					break
				} else {
					log.Errorf(fmt.Errorf("custom error: %v", customErr.Error()).Error())
				}
			}

			if errors.Is(err, AsyncStatus3Error) {
				disconnect()

				break
			} else {
				log.Errorf("unknown error %s :%v", reflect.TypeOf(err).String(), err.Error())
			}
			errCount++
			if errCount > 4 {
				break
			}

			continue
		} else {
			errCount = 0
		}

		msgWG.Wait()
		isWrited = false

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
		go parseData(buff)
		buff = nil
		log.Debugf("release chan")

		msgWG.Done()
	} else {
		buff = append(buff, buf...)
		//log.Debugf("end WTF: %s %d", hex.EncodeToString(buf), len(buf))
	}
}

func parseData(data []byte) {
	if mods.IsValid(data) {
		log.Debugf("data: %s %d", hex.EncodeToString(data), len(data))
		if data[1] == 0x03 && len(data) >= 26 {

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

			influx.PushTo(bmsData)
		}
	}
}
