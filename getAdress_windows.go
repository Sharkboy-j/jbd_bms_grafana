//go:build windows

package main

import (
	"tinygo.org/x/bluetooth"
)

func getAdress(address string) bluetooth.Address {
	mac, err := bluetooth.ParseMAC(address)
	if err != nil {
		log.Errorf("Error parsing UUID address: %v", err)
	}

	return bluetooth.Address{MACAddress: bluetooth.MACAddress{MAC: mac}}
}
