package database

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Sensor struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	DeviceId  string    `json:"device_id"`
	IsDeleted bool      `json:"is_deleted"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at,string"`
	UpdatedAt time.Time `json:"updated_at,string"`
	db        *Database `json:"-"`
	device    *Device   `json:"-"`
}

// TODO
//  - Insert Measurements With Timestamp
//

func (self *Sensor) ImportMeasurementAtLocation(location_id, key string, value float64) error {
	if !self.IsActive {
		return errors.New("sensor is deactivated")
	}

	return self.db.Insert(`
		INSERT INTO measurements (sensor_id, location_id, key, value)
			VALUES ($1, $2, $3, $4)`, self.Id, location_id, key, value)
}

func (self *Sensor) ImportMeasurement(key string, value float64) error {
	if !self.IsActive {
		return errors.New("sensor is deactivated")
	}

	return self.db.Insert(`
		INSERT INTO measurements (sensor_id, key, value)
			VALUES ($1, $2, $3)`, self.Id, key, value)
}

func (self *Sensor) ImportMeasurementsAtLocation(location_id string, data map[string]float64) error {
	if !self.IsActive {
		return errors.New("sensor is deactivated")
	}

	sqlStr := `INSERT INTO measurements (sensor_id, location_id, key, value) VALUES `
	values := []interface{}{}

	for k, v := range data {
		sqlStr += "(?,?,?,?),"
		values = append(values, self.Id, location_id, k, v)
	}

	sqlStr = strings.TrimSuffix(sqlStr, ",")
	count := strings.Count(sqlStr, "?")
	for m := 1; m <= count; m++ {
		sqlStr = strings.Replace(sqlStr, "?", fmt.Sprintf("$%v", m), 1)
	}

	return self.db.Insert(sqlStr, values...)
}

func (self *Sensor) ImportMeasurements(data map[string]float64) error {
	if !self.IsActive {
		return errors.New("sensor is deactivated")
	}

	sqlStr := `INSERT INTO measurements (sensor_id, key, value) VALUES `
	values := []interface{}{}

	for k, v := range data {
		sqlStr += "(?,?,?),"
		values = append(values, self.Id, k, v)
	}

	sqlStr = strings.TrimSuffix(sqlStr, ",")
	count := strings.Count(sqlStr, "?")
	for m := 1; m <= count; m++ {
		sqlStr = strings.Replace(sqlStr, "?", fmt.Sprintf("$%v", m), 1)
	}

	return self.db.Insert(sqlStr, values...)
}

// SetName
func (self *Sensor) SetName(dname string) error {
	self.Name = dname
	return self.Update()
}

// SetType
func (self *Sensor) SetType(dtype string) error {
	self.Type = dtype
	return self.Update()
}

// Delete deletes user
func (self *Sensor) Delete() error {
	self.IsDeleted = true
	return self.Update()
}

// Activate a sensor
func (self *Sensor) Activate() error {
	self.IsActive = true
	return self.Update()
}

// Deactivate a sensor
func (self *Sensor) Deactivate() error {
	self.IsActive = false
	return self.Update()
}

// Update updates user data in database
func (self *Sensor) Update() error {
	return self.db.Insert(`
		UPDATE sensors
			SET
				name=$1,
				type=$2,
				is_deleted=$3,
				is_active=$4
			WHERE id=$5;`,
		self.Name, self.Type, self.IsDeleted, self.IsActive, self.Id)
}
