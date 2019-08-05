package database

import (
	"time"
)

type Sensor struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	DeviceId  string    `json:"device_id"`
	IsDeleted bool      `json:"is_deleted"`
	CreatedAt time.Time `json:"created_at,string"`
	UpdatedAt time.Time `json:"updated_at,string"`
	db        *Database `json:"-"`
	device    *Device   `json:"-"`
}

func (self *Sensor) RecordMeasurement(location_id, key string, value float64) error {
	return self.db.Insert(`
		INSERT INTO measurements (sensor_id, location_id, key, value)
			VALUES ($1, $2, $3, $4)`, self.Id, location_id, key, value)

	// HANDLED BY DATABASE TRIGGER
	// err := self.db.Insert(`
	// 	INSERT INTO measurements (sensor_id, location_id, key, value)
	// 		VALUES ($1, $2, $3, $4)`, self.Id, location_id, key, value)

	// if nil != err {
	// 	return err
	// }

	// add location history
	// return self.device.SetLocation(location_id)
}

// SetPassword sets password
func (self *Sensor) SetName(dname string) error {
	self.Name = dname
	return self.Update()
}

func (self *Sensor) SetType(dtype string) error {
	self.Type = dtype
	return self.Update()
}

// Delete deletes user
func (self *Sensor) Delete() error {
	self.IsDeleted = true
	return self.Update()
}

// Update updates user data in database
func (self *Sensor) Update() error {
	return self.db.Insert(`
		UPDATE sensors
			SET
				name=$1,
				type=$2,
				is_deleted=$3
			WHERE id=$4;`,
		self.Name, self.Type, self.IsDeleted, self.Id)
}
