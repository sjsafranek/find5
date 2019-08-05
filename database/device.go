package database

import (
	"time"
)

type Device struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Username  string    `json:"username"`
	IsDeleted bool      `json:"is_deleted"`
	CreatedAt time.Time `json:"created_at,string"`
	UpdatedAt time.Time `json:"updated_at,string"`
	Sensors   []Sensor  `json:"sensors"`
	db        *Database `json:"-"`
	user      *User     `json:"-"`
}

func (self *Device) CreateSensor(sname, stype string) error {
	return self.db.Insert(`
		INSERT INTO sensors (device_id, name, type)
			VALUES ($1, $2, $3)`, self.Id, sname, stype)
}

func (self *Device) SetLocation(location_id string) error {
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
	return self.Update()
}

// Update updates user data in database
func (self *Device) Update() error {
	return self.db.Insert(`
		UPDATE devices
			SET
				name=$1,
				type=$2,
				is_deleted=$3
			WHERE id=$4;`,
		self.Name, self.Type, self.IsDeleted, self.Id)
}
