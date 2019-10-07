package database

import (
	"time"

	"github.com/paulmach/orb/geojson"
)

type LocationMeasurements struct {
	LocationId         string                `json:"location_id"`
	SensorMeasurements []*SensorMeasurements `json:"sensors"`
}

type SensorMeasurements struct {
	DeviceId           string                `json:"device_id"`
	SensorId           string                `json:"sensor_id"`
	BucketMeasurements []*BucketMeasurements `json:"buckets"`
}

type BucketMeasurements struct {
	BucketId     int64          `json:"bucket_id"`
	Measurements []*Measurement `json:"measurements"`
}

type Measurement struct {
	Key   string  `json:"key"`
	Value float64 `json:"value"`
}

type DeviceLocations struct {
	LocationId   string                 `json:"location_id"`
	LocationName string                 `json:"location_name"`
	Geometry     geojson.Geometry       `json:"geometry"`
	Devices      []*DeviceLocationStats `json:"devices"`
}

type DeviceLocationStats struct {
	DeviceId           string                 `json:"device_id"`
	DeviceName         string                 `json:"device_name"`
	FirstTimestamp     time.Time              `json:"first_timestamp,string"`
	LastestTimestamp   time.Time              `json:"latest_timestamp,string"`
	AverageProbability float64                `json:"average_probability"`
	Sensors            []*SensorLocationStats `json:"sensors"`
}

type SensorLocationStats struct {
	SensorId     string `json:"sensor_id"`
	SensorName   string `json:"sensor_name"`
	Measurements int    `json:"measurements"`
}

type MeasurementLocations struct {
	LocationId   string                      `json:"location_id"`
	LocationName string                      `json:"location_name"`
	Geometry     geojson.Geometry            `json:"geometry"`
	Scanners     []*MeasurementLocationStats `json:"scanners"`
}

type MeasurementLocationStats struct {
	Key              string    `json:"key"`
	Sensors          int       `json:"sensors"`
	Count            int       `json:"count"`
	Min              float64   `json:"min"`
	Max              float64   `json:"max"`
	Stddev           float64   `json:"stddev"`
	Mean             float64   `json:"mean"`
	FirstTimestamp   time.Time `json:"first_timestamp,string"`
	LastestTimestamp time.Time `json:"latest_timestamp,string"`
}
