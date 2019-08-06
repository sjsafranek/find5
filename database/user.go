package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/paulmach/orb/geojson"
)

type User struct {
	Username    string    `json:"username"`
	Password    string    `json:"-"`
	Email       string    `json:"email"`
	Apikey      string    `json:"apikey,omitempty"`
	SecretToken string    `json:"secret_token,omitempty"`
	IsDeleted   bool      `json:"is_deleted"`
	CreatedAt   time.Time `json:"created_at,string"`
	UpdatedAt   time.Time `json:"updated_at,string"`
	db          *Database `json:"-"`
}

// SetPassword sets password
func (self *User) SetPassword(password string) error {
	self.Password = password
	return self.Update()
}

// Delete deletes user
func (self *User) Delete() error {
	self.IsDeleted = true
	return self.Update()
}

// Update updates user data in database
func (self *User) Update() error {
	return self.db.Insert(`
		UPDATE users
			SET
				email=$1,
				password=$2,
				is_deleted=$3
			WHERE username=$4;`, self.Email, self.Password, self.IsDeleted, self.Username)
}

func (self *User) Marshal() (string, error) {
	b, err := json.Marshal(self)
	if nil != err {
		return "", err
	}
	return string(b), err
}

func (self *User) Unmarshal(data string) error {
	return json.Unmarshal([]byte(data), self)
}

// IsPassword checks if provided password/hash matches database record
func (self *User) IsPassword(password string) (bool, error) {
	match := false
	return match, self.db.Exec(func(conn *sql.DB) error {
		rows, err := conn.Query(`
		SELECT
			-- back door for using hashed password for login
			CASE
				WHEN password = $2 THEN TRUE
				ELSE password = crypt($2, salt)
			END AS matched
			-- hash = crypt($2, salt) AS matched
		FROM users
		WHERE username=$1;`, self.Username, password)

		if nil != err {
			return err
		}

		for rows.Next() {
			rows.Scan(&match)
			return nil
		}

		return errors.New("Not found")
	})
}

func (self *User) CreateDevice(dname, dtype string) (*Device, error) {
	err := self.db.Insert(`
		INSERT INTO devices(username, name, type)
			VALUES ($1, $2, $3);`, self.Username, dname, dtype)

	if nil != err {
		return &Device{}, err
	}

	return self.GetDeviceByName(dname)
}

func (self *User) GetDevices() ([]*Device, error) {
	var devices []*Device
	return devices, self.db.Exec(func(conn *sql.DB) error {

		rows, err := conn.Query(`
		SELECT json_agg(d) FROM (
			SELECT
				devices.id,
				devices.name,
				devices.type,
				devices.username,
				devices.is_deleted,
				to_char(devices.created_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"') as created_at,
		        to_char(devices.updated_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"') as updated_at,
				(
					SELECT json_agg(s) FROM (
						SELECT
							sensors.id,
							sensors.name,
							sensors.type,
							sensors.device_id,
							sensors.is_deleted,
							to_char(sensors.created_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"') as created_at,
							to_char(sensors.updated_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"') as updated_at
						FROM sensors
						WHERE device_id=devices.id
						AND sensors.is_deleted = false
					) s
				) AS sensors
			FROM devices
			WHERE devices.username = $1
			AND devices.is_deleted = false
		) d;`, self.Username)

		if nil != err {
			return err
		}

		for rows.Next() {
			var temp string
			rows.Scan(&temp)
			err = json.Unmarshal([]byte(temp), &devices)
			if nil != err {
				return err
			}
		}

		// add database to objects
		for i := range devices {
			devices[i].db = self.db
			devices[i].user = self
			if nil != devices[i].Sensors {
				for j := range devices[i].Sensors {
					devices[i].Sensors[j].db = self.db
					devices[i].Sensors[j].device = devices[i]
				}
			}
		}

		return nil
	})
}

func (self *User) GetDeviceByName(device_name string) (*Device, error) {
	devices, err := self.GetDevices()
	if nil != err {
		return &Device{}, err
	}
	for _, device := range devices {
		if device.Name == device_name {
			return device, nil
		}
	}
	return &Device{}, errors.New("Device not found")
}
func (self *User) GetDeviceById(device_id string) (*Device, error) {
	devices, err := self.GetDevices()
	if nil != err {
		return &Device{}, err
	}
	for _, device := range devices {
		if device.Id == device_id {
			return device, nil
		}
	}
	return &Device{}, errors.New("Not found")
}

func (self *User) CreateLocation(lname string, lng, lat float64) error {
	gjson := fmt.Sprintf(`{"type":"Point","coordinates":[%v,%v]}`, lng, lat)
	return self.db.Insert(`
		INSERT INTO locations(username, name, geom)
			VALUES ($1, $2, ST_SetSRID(ST_GeomFromGeoJSON($3),4326));
		`, self.Username, lname, gjson)
}

func (self *User) GetLocations() (*geojson.FeatureCollection, error) {
	var layer geojson.FeatureCollection
	return &layer, self.db.Exec(func(conn *sql.DB) error {

		rows, err := conn.Query(`
		SELECT
			json_build_object(
				'type', 'FeatureCollection',
				'features', json_agg(c)
			) AS geojson
		FROM (
			SELECT
				'Feature' AS type,
				ST_AsGeoJSON(geom)::jsonb AS geometry,
				json_build_object(
					'id', id,
					'name', name,
					'username', username,
					'created_at', to_char(locations.created_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"'),
					'updated_at', to_char(locations.updated_at, 'YYYY-MM-DD"T"HH:MI:SS"Z"')
				) AS properties
			FROM locations
			WHERE
					geom IS NOT NULL
				AND
					locations.username = $1
				AND
					locations.is_deleted = false
		) c;`, self.Username)

		if nil != err {
			return err
		}

		for rows.Next() {
			var temp string
			rows.Scan(&temp)
			err = json.Unmarshal([]byte(temp), &layer)
			if nil != err {
				return err
			}
		}

		return nil
	})
}
