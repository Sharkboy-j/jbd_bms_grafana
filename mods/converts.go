package mods

import (
	"encoding/binary"
	"fmt"
)

func GetMOS(byteValue byte) MOSStatus {
	charging := (byteValue & 1) != 0
	discharging := ((byteValue >> 1) & 1) != 0

	return MOSStatus{
		Charging:    charging,
		Discharging: discharging,
	}
}

func ToFloat(data []byte) float32 {
	intValue := binary.BigEndian.Uint16(data)
	floatValue := float32(intValue)

	return floatValue
}

func ToInt(data byte) int32 {
	int32Value := int32(data)

	return int32Value
}

func ToVersion(versionByte byte) string {
	major := versionByte >> 4   // старшие 4 бита
	minor := versionByte & 0x0F // младшие 4 бита

	version := fmt.Sprintf("%d.%d", major, minor)

	return version
}

func ToPercents(byteValue byte) int {
	intValue := int(byteValue)

	return intValue
}
