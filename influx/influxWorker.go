package influx

import (
	"bleTest/logger"
	"bleTest/mods"
	"context"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"os"
	"strconv"
	"time"
)

var (
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
	log      *logger.Logger
)

func Init(logger *logger.Logger) {
	log = logger
	influxDBURL := os.Getenv("INFLUX_DBURL")
	token := os.Getenv("INFLUX_TOKEN")
	org := os.Getenv("INFLUX_ORG")
	bucket := os.Getenv("INFLUX_BUCKET")

	client = influxdb2.NewClient(influxDBURL, token)
	writeAPI = client.WriteAPIBlocking(org, bucket)
}

func PushData(data *mods.JbdData) {
	p := influxdb2.NewPointWithMeasurement("jbd_data").
		AddField("current", data.Current).
		AddField("volts", data.Volts).
		AddField("capacity", data.RemainingCapacity).
		AddField("perc", data.RemainingPercent).
		AddField("leftTime", data.Left)

	for i, v := range data.Temp {
		p.AddField("temp"+strconv.Itoa(i), v)
	}

	p.SetTime(time.Now())

	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		log.Errorf("Error writing point to InfluxDB: %v", err)
	}
}

func PushCells(data *mods.JbdData) {
	p := influxdb2.NewPointWithMeasurement("jbd_data")

	for i, v := range data.Cells {
		p.AddField("cell"+strconv.Itoa(i), v)
	}

	p.AddField("maxCell", data.MaxCell).
		AddField("minCell", data.MinCell).
		AddField("diff", data.Diff).
		AddField("avg", data.Avg).
		SetTime(time.Now())

	if err := writeAPI.WritePoint(context.Background(), p); err != nil {
		log.Errorf("Error writing point to InfluxDB: %v", err)
	}
}
