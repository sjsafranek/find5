package ai

import (
	"fmt"

	"github.com/sjsafranek/find5/lib/database"
	"github.com/sjsafranek/ligneous"
)

var (
	logger = ligneous.AddLogger("ai", "trace", "./log/find5")
)

func New() *AI {
	return &AI{}
}

// SensorData is the typical data structure for storing sensor data.
type SensorData struct {
	// Timestamp is the unique identifier, the time in milliseconds
	Timestamp int64 `json:"t"`
	// Family is a group of devices
	Family string `json:"f"`
	// Device are unique within a family
	Device string `json:"d"`
	// Location is optional, used for classification
	Location string `json:"l,omitempty"`
	// Sensors contains a map of map of sensor data
	Sensors map[string]map[string]interface{} `json:"s"`
	// GPS is optional
	// GPS GPS `json:"gps,omitempty"`
}

type AI struct{}

// HACK
func (self *AI) convertMeasurementsToSensorData(locationMeasurements []*database.LocationMeasurements, family string) []SensorData {

	datas := []SensorData{}

	locationTimeBucketData := make(map[string]SensorData)

	for _, location := range locationMeasurements {
		for _, sensor := range location.SensorMeasurements {
			for _, bucket := range sensor.BucketMeasurements {

				hsh := fmt.Sprintf("%v-%v", location.LocationId, bucket.BucketId)

				// if location already in map
				if _, ok := locationTimeBucketData[hsh]; ok {
					sd := SensorData{
						Family:    family,
						Device:    sensor.DeviceId,
						Location:  location.LocationId,
						Timestamp: bucket.BucketId,
						Sensors:   make(map[string]map[string]interface{}),
					}
					sd.Sensors[sensor.SensorId] = make(map[string]interface{})
					for _, measurement := range bucket.Measurements {
						locationTimeBucketData[hsh].Sensors[sensor.SensorId][measurement.Key] = measurement.Value
					}
					continue
				}

				// create new locdation bucket
				sd := SensorData{
					Family:    family,
					Device:    sensor.DeviceId,
					Location:  location.LocationId,
					Timestamp: bucket.BucketId,
					Sensors:   make(map[string]map[string]interface{}),
				}
				sd.Sensors[sensor.SensorId] = make(map[string]interface{})
				for _, measurement := range bucket.Measurements {
					sd.Sensors[sensor.SensorId][measurement.Key] = measurement.Value
				}
				locationTimeBucketData[hsh] = sd

			}
		}
	}

	for i := range locationTimeBucketData {
		datas = append(datas, locationTimeBucketData[i])
	}

	return datas
}

//.end

func (self *AI) Calibrate(locationMeasurements []*database.LocationMeasurements, family string, crossValidation ...bool) {
	datas := self.convertMeasurementsToSensorData(locationMeasurements, family)
	fmt.Println(datas)
}
