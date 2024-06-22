//go:build darwin

package bluetoothHelper

import (
	"tinygo.org/x/bluetooth"
)

func GetAdress(log *logger.Logger, address string) *bluetooth.Address {
	uuid, err := bluetooth.ParseUUID(address)
	if err != nil {
		log.Errorf("Error parsing UUID address: %v", err)
	}

	return &bluetooth.Address{UUID: uuid}
}
