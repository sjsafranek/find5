package api

import (
	"encoding/json"
)

type Query struct {
	Method     string                        `json:"method"`
	Email      string                        `json:"email"`
	Username   string                        `json:"username"`
	Password   string                        `json:"password"`
	Apikey     string                        `json:"apikey"`
	DeviceId   string                        `json:"device_id"`
	LocationId string                        `json:"location_id"`
	Name       string                        `json:"name"`
	Type       string                        `json:"type"`
	Data       map[string]map[string]float64 `json:"data"`
}

func (self *Query) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), self)
}
