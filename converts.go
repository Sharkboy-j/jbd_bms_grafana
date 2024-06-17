package main

import (
	"encoding/binary"
	"fmt"
)

func getMOS(byteValue byte) MOSStatus {
	charging := (byteValue & 1) != 0
	discharging := ((byteValue >> 1) & 1) != 0

	return MOSStatus{
		Charging:    charging,
		Discharging: discharging,
	}
}

func toFloat(data []byte) float32 {
	intValue := binary.BigEndian.Uint16(data)
	floatValue := float32(intValue)

	return floatValue
}

func toInt(data byte) int32 {
	int32Value := int32(data)

	return int32Value
}

func toVersion(versionByte byte) string {
	major := versionByte >> 4   // старшие 4 бита
	minor := versionByte & 0x0F // младшие 4 бита

	// Форматирование строки версии
	version := fmt.Sprintf("%d.%d", major, minor)

	return version
}

func toPercents(byteValue byte) int {
	intValue := int(byteValue)

	return intValue
}
