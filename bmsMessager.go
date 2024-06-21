package main

import (
	"bleTest/app"
	"bleTest/mods"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/godbus/dbus/v5"
	"reflect"
	"time"
)

func writerChan() {
	errCount := 0
	log.Debugf("start write cycle")

	for {
		if app.Canceled {
			break
		}

		resp, err := txChars.WriteWithoutResponse(ReadMessage)
		log.Debugf("writed")
		msgWG.Add(1)

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

			log.Errorf("unknown error %s :%v", reflect.TypeOf(err).String(), err.Error())
			errCount++
		} else {
			errCount = 0
		}

		if errCount > 4 {
			break
		}

		log.Debugf("wait fo response")
		msgWG.Wait()

		time.Sleep(3 * time.Second)
	}
}

func recCallb(buf []byte) {
	if buf[0] == StartBit {
		buff = buf
		log.Debugf("start WTF: %s %d", hex.EncodeToString(buf), len(buf))

	} else if buf[len(buf)-1] == StopBit {
		log.Debugf("body WTF: %s %d", hex.EncodeToString(buf), len(buf))

		buff = append(buff, buf...)
		log.Debugf("read end")
		go parseData()
		buff = nil
		log.Debugf("release chan")
		msgWG.Done()
	} else {
		log.Debugf("end WTF: %s %d", hex.EncodeToString(buf), len(buf))
		buff = append(buff, buf...)
	}
}

func parseData() {
	if mods.IsValid(buff) {
		log.Debugf("data: %s %d", hex.EncodeToString(buff), len(buff))
		if buff[1] == 0x03 && len(buff) >= 26 {

			mos := mods.GetMOS(buff[24])
			bmsData.Volts = mods.ToFloat([]byte{buff[4], buff[5]}) / 100
			bmsData.Current = mods.ToFloat([]byte{buff[6], buff[7]}) / 100
			bmsData.RemainingCapacity = mods.ToFloat([]byte{buff[8], buff[9]}) / 100
			bmsData.NominalCapcity = mods.ToFloat([]byte{buff[10], buff[11]}) / 100
			bmsData.Cycles = mods.ToFloat([]byte{buff[12], buff[13]})
			bmsData.Version = mods.ToVersion(buff[22])
			bmsData.RemainingPercent = mods.ToPercents(buff[23])
			bmsData.Series = mods.ToInt(buff[25])
			//Temp:              toFloat(),
			bmsData.MosChargingEnabled = mos.Charging
			bmsData.MosDischargingEnabled = mos.Discharging

			temp := make([]float32, int(buff[26]))
			for i := 0; i < int(buff[26]); i++ {
				temperature := (float32(buff[27+i*2])*256 + float32(buff[28+i*2]) - 2731) / 10
				temp = append(temp, temperature)
			}
			bmsData.Temp = temp

			pushTo(bmsData)
		}
	}
}
