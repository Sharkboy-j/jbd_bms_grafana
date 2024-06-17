package main

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"log"
	"time"
)

var client influxdb2.Client
var writeAPI api.WriteAPIBlocking

func initData() {
	// Define your InfluxDB server URL and authentication token
	const influxDBURL = "http://10.0.0.196:8086"
	const token = "ndAzO_IU75cmGIEseZMEE9ihCYHIxn7qDkvcNlrUcw2ajWgmmt9VKcdlgFsPN8O-_FDga3kEtLnUYl8wHskVKw=="
	const org = "jbd"
	const bucket = "jbd"

	// Create a new client
	client = influxdb2.NewClient(influxDBURL, token)
	writeAPI = client.WriteAPIBlocking(org, bucket)
}

func pushTo(data *JbdData) {
	// Create a point using the struct fields
	p := influxdb2.NewPointWithMeasurement("jbd_data").
		AddField("current", data.Current).
		AddField("volts", data.Volts).
		AddField("capacity", data.RemainingCapacity).
		AddField("perc", data.RemainingPercent).
		SetTime(time.Now())

	// Write the point to InfluxDB
	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		log.Fatalf("Error writing point to InfluxDB: %v", err)
	}

	fmt.Println("Data written successfully!")
}
