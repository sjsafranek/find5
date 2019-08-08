package database

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
