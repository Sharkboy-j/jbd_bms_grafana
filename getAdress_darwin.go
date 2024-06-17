//go:build darwin

package main

import (
	"log"
	"tinygo.org/x/bluetooth"
)

func getAdress(address string) bluetooth.Address {
	uuid, err := bluetooth.ParseUUID(address)
	if err != nil {
		log.Fatalf("Error parsing UUID address: %v", err)
	}

	return bluetooth.Address{UUID: uuid}
}
