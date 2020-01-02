package api

import (
	"encoding/json"
	"time"

	"github.com/paulmach/orb/geojson"
	"github.com/sjsafranek/find5/findapi/lib/database"
)

const (
	VERSION = "5.0.1"
)

type Request struct {
	Method  string `json:"method,omitempty"`
	Version string `json:"version"`
	Params  Params `json:"params"`
	Id      string `json:"id,ompitempty"`
}

type Params struct {
	Email      string                        `json:"email,omitempty"`
	Username   string                        `json:"username,omitempty"`
	Password   string                        `json:"password,omitempty"`
	Apikey     string                        `json:"apikey,omitempty"`
	DeviceId   string                        `json:"device_id,omitempty"`
	SensorId   string                        `json:"sensor_id,omitempty"`
	LocationId string                        `json:"location_id,omitempty"`
	Latitude   float64                       `json:"latitude,omitempty"`
	Longitude  float64                       `json:"longitude,omitempty"`
	Name       string                        `json:"name,omitempty"`
	Type       string                        `json:"type,omitempty"`
	Data       map[string]map[string]float64 `json:"data,omitempty"`
	Timestamp  time.Time                     `json:"timestamp,string"`
}

func (self *Request) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), self)
}

type ResponseData struct {
	Users                []*database.User                 `json:"users,omitempty"`
	User                 *database.User                   `json:"user,omitempty"`
	Devices              []*database.Device               `json:"devices,omitempty"`
	Device               *database.Device                 `json:"device,omitempty"`
	Sensors              []*database.Sensor               `json:"sensors,omitempty"`
	Sensor               *database.Sensor                 `json:"sensor,omitempty"`
	Locations            *geojson.FeatureCollection       `json:"locations,omitempty"`
	Measurements         []*database.LocationMeasurements `json:"measurements,omitempty"`
	DeviceLocations      []*database.DeviceLocations      `json:"device_locations,omitempty"`
	MeasurementLocations []*database.MeasurementLocations `json:"measurements_locations,omitempty"`
}

type Response struct {
	Id      string       `json:"id"`
	Version string       `json:"version"`
	Status  string       `json:"status"`
	Message string       `json:"message,omitempty"`
	Error   string       `json:"error,omitempty"`
	Data    ResponseData `json:"data,omitempty"`
}

func (self *Response) Marshal() (string, error) {
	b, err := json.Marshal(self)
	if nil != err {
		return "", err
	}
	return string(b), err
}

func (self *Response) SetError(err error) {
	self.Status = "error"
	self.Error = err.Error()
}
