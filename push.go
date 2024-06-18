package main

import (
	"context"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"time"
)

var client influxdb2.Client
var writeAPI api.WriteAPIBlocking

func initData() {
	const influxDBURL = "http://10.0.0.196:8086"
	const token = "ndAzO_IU75cmGIEseZMEE9ihCYHIxn7qDkvcNlrUcw2ajWgmmt9VKcdlgFsPN8O-_FDga3kEtLnUYl8wHskVKw=="
	const org = "jbd"
	const bucket = "jbd"

	client = influxdb2.NewClient(influxDBURL, token)
	writeAPI = client.WriteAPIBlocking(org, bucket)
}

func pushTo(data *JbdData) {
	p := influxdb2.NewPointWithMeasurement("jbd_data").
		AddField("current", data.Current).
		AddField("volts", data.Volts).
		AddField("capacity", data.RemainingCapacity).
		AddField("perc", data.RemainingPercent).
		SetTime(time.Now())

	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		log.Errorf("Error writing point to InfluxDB: %v", err)
	}

	log.Infof("Data written successfully!")
}
