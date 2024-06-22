//go:build !darwin

package bluetoothHelper

import (
	"bleTest/logger"
	"tinygo.org/x/bluetooth"
)

func GetAdress(log *logger.Logger, address string) *bluetooth.Address {
	mac, err := bluetooth.ParseMAC(address)
	if err != nil {
		log.Errorf("Error parsing UUID address: %v", err)
	}

	return &bluetooth.Address{MACAddress: bluetooth.MACAddress{MAC: mac}}
}
