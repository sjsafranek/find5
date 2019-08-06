package api

import (
	"encoding/json"

	"github.com/paulmach/orb/geojson"
	"github.com/sjsafranek/find5/database"
)

type Request struct {
	Method     string                        `json:"method"`
	Email      string                        `json:"email"`
	Username   string                        `json:"username"`
	Password   string                        `json:"password"`
	Apikey     string                        `json:"apikey"`
	DeviceId   string                        `json:"device_id"`
	SensorId   string                        `json:"sensor_id"`
	LocationId string                        `json:"location_id"`
	Latitude   float64                       `json:"latitude"`
	Longitude  float64                       `json:"longitude"`
	Name       string                        `json:"name"`
	Type       string                        `json:"type"`
	Data       map[string]map[string]float64 `json:"data"`
}

func (self *Request) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), self)
}

type ResponseData struct {
	Users        []*database.User                 `json:"users,omitempty"`
	User         *database.User                   `json:"user,omitempty"`
	Devices      []*database.Device               `json:"devices,omitempty"`
	Device       *database.Device                 `json:"device,omitempty"`
	Sensors      []*database.Sensor               `json:"sensors,omitempty"`
	Sensor       *database.Sensor                 `json:"sensor,omitempty"`
	Locations    *geojson.FeatureCollection       `json:"locations,omitempty"`
	Measurements []*database.LocationMeasurements `json:"measurements,omitempty"`
}

type Response struct {
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
