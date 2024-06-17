package main

import "fmt"

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
	MosChargingEnabled    bool
	MosDischargingEnabled bool
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
