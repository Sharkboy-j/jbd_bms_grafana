package mods

import (
	"fmt"
)

const StartBit byte = 0xDD
const StopBit byte = 0x77

type MOSStatus struct {
	Charging    bool
	Discharging bool
}

type JbdData struct {
	Volts             float32
	Current           float32
	NominalCapcity    float32
	RemainingCapacity float32
	Cycles            float32

	Version               string
	RemainingPercent      int
	Series                int32
	Temp                  []float32
	Cells                 []float32
	Diff                  float32
	MinCell               float32
	MaxCell               float32
	MosChargingEnabled    bool
	MosDischargingEnabled bool
}

func (b JbdData) GetMaxMin() (max float32, min float32) {
	if len(b.Cells) == 0 {
		return 0, 0 // return NaN if the slice is empty
	}

	max = b.Cells[0]
	min = b.Cells[0]

	for _, value := range b.Cells {
		if value > max {
			max = value
		}
		if value < min {
			min = value
		}
	}

	return max, min
}

func IsValid(buff []byte) bool {
	return buff[0] == StartBit && buff[len(buff)-1] == StopBit
}

func (b JbdData) String() string {
	var temp string

	for _, v := range b.Temp {
		temp += fmt.Sprintf("%0.2f ", v)
	}

	return fmt.Sprintf("Volts: %f\n Current: %f\n Capcity: %f\n RemainingCapacity: %f\n Cycles: %f\n Version: %s\n RemainingPercent: %d\n Series: %d\n Charging: %t\n Discharging: %t\n Temps: %s\n",
		b.Volts,
		b.Current,
		b.NominalCapcity,
		b.RemainingCapacity,
		b.Cycles,
		b.Version,
		b.RemainingPercent,
		b.Series,
		b.MosChargingEnabled,
		b.MosDischargingEnabled,
		temp)
}
