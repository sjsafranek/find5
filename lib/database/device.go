package database

import (
	"encoding/json"
	"errors"
	"time"
)

type Device struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Username  string    `json:"username"`
	IsDeleted bool      `json:"is_deleted"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at,string"`
	UpdatedAt time.Time `json:"updated_at,string"`
	Sensors   []*Sensor `json:"sensors"`
	db        *Database `json:"-"`
	user      *User     `json:"-"`
}

func (self *Device) CreateSensor(sname, stype string) error {
	if !self.IsActive {
		return errors.New("device is deactivated")
	}

	return self.db.Insert(`
		INSERT INTO sensors (device_id, name, type)
			VALUES ($1, $2, $3)`, self.Id, sname, stype)
}

func (self *Device) SetLocation(location_id string) error {
	if !self.IsActive {
		return errors.New("device is deactivated")
	}

	return self.db.Insert(`
		INSERT INTO location_history (device_id, location_id)
			VALUES ($1, $2)`, self.Id, location_id)
}

// SetPassword sets password
func (self *Device) SetName(dname string) error {
	self.Name = dname
	return self.Update()
}

func (self *Device) SetType(dtype string) error {
	self.Type = dtype
	return self.Update()
}

// Delete deletes user
func (self *Device) Delete() error {
	self.IsDeleted = true
	err := self.Update()
	if nil != err {
		return err
	}
	// TODO move this logic into database
	// cascade "delete" all the sensors
	for _, sensor := range self.Sensors {
		err = sensor.Delete()
		if nil != err {
			return err
		}
	}
	return nil
}

// Activate a device
func (self *Device) Activate() error {
	self.IsActive = true
	return self.Update()
}

// Deactivate a device
func (self *Device) Deactivate() error {
	self.IsActive = false
	return self.Update()
}

// Update updates user data in database
func (self *Device) Update() error {
	return self.db.Insert(`
		UPDATE devices
			SET
				name=$1,
				type=$2,
				is_deleted=$3,
				is_active=$4
			WHERE id=$5;`,
		self.Name, self.Type, self.IsDeleted, self.IsActive, self.Id)
}

func (self *Device) Marshal() (string, error) {
	b, err := json.Marshal(self)
	if nil != err {
		return "", err
	}
	return string(b), err
}

func (self *Device) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), self)
}

func (self *Device) GetSensors() ([]*Sensor, error) {
	return self.Sensors, nil
}

func (self *Device) GetSensorIds() ([]string, error) {
	sensorIds := []string{}
	for _, sensor := range self.Sensors {
		sensorIds = append(sensorIds, sensor.Id)
	}
	return sensorIds, nil
}

func (self *Device) GetSensorByName(sensor_name string) (*Sensor, error) {
	for _, sensor := range self.Sensors {
		if sensor.Name == sensor_name {
			return sensor, nil
		}
	}
	return &Sensor{}, errors.New("sensor not found")
}
func (self *Device) GetSensorById(sensor_id string) (*Sensor, error) {
	for _, sensor := range self.Sensors {
		if sensor.Id == sensor_id {
			return sensor, nil
		}
	}
	return &Sensor{}, errors.New("sensor not found")
}
