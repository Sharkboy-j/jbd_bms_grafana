//go:build windows

package main

import (
	"log"
	"tinygo.org/x/bluetooth"
)

func getAdress(address string) bluetooth.Address {
	mac, err := bluetooth.ParseMAC(address)
	if err != nil {
		log.Fatalf("Error parsing MAC address: %v", err)
	}

	return bluetooth.Address{MACAddress: bluetooth.MACAddress{MAC: mac}}
}
