package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/paulmach/orb/geojson"
	// "github.com/sjsafranek/logger"
)

type User struct {
	Username    string    `json:"username"`
	Password    string    `json:"-"`
	Email       string    `json:"email"`
	Apikey      string    `json:"apikey,omitempty"`
	SecretToken string    `json:"secret_token,omitempty"`
	IsDeleted   bool      `json:"is_deleted"`
	IsActive    bool      `json:"is_active"`
	IsSuperuser bool	  `json:"is_superuser"`
	CreatedAt   time.Time `json:"created_at,string"`
	UpdatedAt   time.Time `json:"updated_at,string"`
	db          *Database `json:"-"`
}

// SetEmail sets user email
func (self *User) SetEmail(email string) error {
	self.Email = email
	return self.Update()
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

// Activate deletes user
func (self *User) Activate() error {
	self.IsActive = true
	return self.Update()
}

// Deactivate deletes user
func (self *User) Deactivate() error {
	self.IsActive = false
	return self.Update()
}

// Update updates user data in database
func (self *User) Update() error {
	return self.db.Insert(`
		UPDATE users
			SET
				email=$1,
				password=$2,
				is_deleted=$3,
				is_active=$4
			WHERE username=$5;`, self.Email, self.Password, self.IsDeleted, self.IsActive, self.Username)
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
		rows, err := conn.Query(`SELECT is_password($1, $2);`, self.Username, password)

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

/**
 * Social Accounts
 */
// CreateSocialAccountIfNotExists
// https://stackoverflow.com/questions/4069718/postgres-insert-if-does-not-exist-already
// ON CONFLICT DO NOTHING/UPDATE
// http://www.postgresqltutorial.com/postgresql-upsert/
func (self *User) CreateSocialAccountIfNotExists(user_id, username, account_type string) error {
	err := self.db.Insert(`
		INSERT INTO social_accounts(id, name, type, email)
			VALUES ($1, $2, $3, $4)
				ON CONFLICT DO NOTHING;
	`, user_id, username, account_type, self.Username)
	if nil != err && strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
		return nil
	}
	return nil
}

/**
 * APP FUNCTIONS
 */
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
		SELECT
		    json_agg(devices_json)
		FROM devices_view
		WHERE username = $1
		AND is_deleted = false
		;`, self.Username)

		if nil != err {
			return checkError(err)
		}

		for rows.Next() {
			var temp string
			rows.Scan(&temp)

			// handle no records being returned
			if "" == temp {
				return nil
			}

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

func (self *User) ExportMeasurements() ([]*LocationMeasurements, error) {
	var locationMeasurments []*LocationMeasurements

	return locationMeasurments, self.db.Exec(func(conn *sql.DB) error {

		rows, err := conn.Query(`
			WITH measurements AS (
			        SELECT
			            (EXTRACT(epoch FROM measurements.created_at) / EXTRACT(epoch FROM INTERVAL '5 sec'))::INTEGER AS bucket,
			            measurements.location_id,
			            sensors.id AS sensor_id,
						devices.id AS device_id,
			            json_agg(
			                json_build_object(
			                    'key',
			                    measurements.key,
			                    'value',
			                    measurements.value
			                )
			            ) AS measurements
			        FROM measurements
			        INNER JOIN sensors
			            ON sensors.id = measurements.sensor_id
						AND sensors.is_deleted = false
			        INNER JOIN devices
			            ON devices.id = sensors.device_id
			            AND devices.username = $1
						AND devices.is_deleted = false
			        WHERE
			            measurements.location_id IS NOT NULL
			        GROUP BY bucket, measurements.location_id, sensors.id, devices.id
			    ),
			    buckets AS (
			        SELECT
			            measurements.location_id,
			            measurements.sensor_id,
			            json_build_object(
			                'bucket_id',
			                measurements.bucket,
			                'measurements',
			                json_agg(measurements.measurements)->0
			            ) AS bucket
			        FROM measurements
			        GROUP BY measurements.bucket, measurements.location_id, measurements.sensor_id
			    ),
			    location_buckets AS (
			        SELECT
			            buckets.sensor_id,
			            buckets.location_id,
			            json_agg(buckets.bucket) AS buckets
			        FROM buckets
			        GROUP BY buckets.location_id, buckets.sensor_id
			    ),
			    sensor_locations AS (
			        SELECT
			            location_buckets.location_id,
			            json_agg(
			                json_build_object(
			                'sensor_id',
			                location_buckets.sensor_id,
			                'buckets',
			                location_buckets.buckets
			                )
			            ) AS sensors
			        FROM location_buckets
			        GROUP BY location_buckets.location_id
			    )

			SELECT json_agg(c) FROM (
			    SELECT
			        location_id,
			        sensors
			    FROM sensor_locations
			) c;`, self.Username)

		if nil != err {
			return err
		}

		for rows.Next() {
			var temp string
			rows.Scan(&temp)
			err = json.Unmarshal([]byte(temp), &locationMeasurments)
			if nil != err {
				return err
			}
		}

		return nil
	})
}

func (self *User) ExportDevicesByLocation() ([]*DeviceLocations, error) {
	var deviceLocations []*DeviceLocations

	return deviceLocations, self.db.Exec(func(conn *sql.DB) error {

		rows, err := conn.Query(`
			WITH
				device_sensor_locations AS (
				    SELECT
				        locations.id AS location_id,
				        locations.name AS location_name,
				        ST_AsGeoJSON(locations.geom)::JSONB AS geometry,
				        devices.id AS device_id,
				        devices.name AS device_name,
				        sensors.id AS sensor_id,
				        sensors.name AS sensor_name,
						to_char(MIN(location_history.created_at), 'YYYY-MM-DD"T"HH:MI:SS"Z"') AS first_timestamp,
						to_char(MAX(location_history.created_at), 'YYYY-MM-DD"T"HH:MI:SS"Z"') AS lastest_timestamp,
				        AVG(location_history.probability) AS average_probability,
				        COUNT(measurements.*) AS num_measurements
				    FROM
				        location_history
				    INNER JOIN devices
				        ON devices.id = location_history.device_id
				        ANd devices.username = $1
				        AND devices.is_deleted = false
				    INNER JOIN locations
				        ON locations.id = location_history.location_id
				        AND locations.is_deleted = false
				    INNER JOIN sensors
				        ON sensors.device_id = devices.id
						AND sensors.is_deleted = false
				    LEFT JOIN measurements
				        ON measurements.sensor_id = sensors.id
				        AND measurements.created_at >= (NOW() - INTERVAL '5 minutes')
				    WHERE
				        location_history.created_at >= (NOW() - INTERVAL '5 minutes')
				    GROUP BY locations.id, devices.id, sensors.id
				),
				device_locations AS (
				    SELECT
				        device_sensor_locations.location_id,
				        device_sensor_locations.location_name,
				        device_sensor_locations.geometry,
				        device_sensor_locations.device_id,
				        device_sensor_locations.device_name,
						device_sensor_locations.first_timestamp,
				        device_sensor_locations.lastest_timestamp,
				        device_sensor_locations.average_probability,
				        json_agg(
				            json_build_object(
				                'sensor_id',
				                device_sensor_locations.sensor_id,
				                'sensor_name',
				                device_sensor_locations.sensor_name,
				                'measurements',
				                device_sensor_locations.num_measurements
				            )
				        ) AS sensors
				    FROM device_sensor_locations
				    GROUP BY
				        device_sensor_locations.location_id,
				        device_sensor_locations.location_name,
				        device_sensor_locations.geometry,
				        device_sensor_locations.device_id,
				        device_sensor_locations.device_name,
				        device_sensor_locations.first_timestamp,
				        device_sensor_locations.lastest_timestamp,
				        device_sensor_locations.average_probability
				)

			SELECT json_agg(c)
			FROM (
			    SELECT
			        device_locations.location_id,
			        device_locations.location_name,
			        device_locations.geometry,
			        json_agg(
			            json_build_object(
			                'device_id',
			                device_locations.device_id,
			                'device_name',
			                device_locations.device_name,
							'first_timestamp',
							device_locations.first_timestamp,
			                'lastest_timestamp',
			                device_locations.lastest_timestamp,
			                'average_probability',
			                device_locations.average_probability,
			                'sensors',
			                device_locations.sensors
			            )
			        ) AS devices
			    FROM device_locations
			    GROUP BY
			        device_locations.location_id,
			        device_locations.location_name,
			        device_locations.geometry
			) c; `, self.Username)

		if nil != err {
			return err
		}

		for rows.Next() {
			var temp string
			rows.Scan(&temp)
			err = json.Unmarshal([]byte(temp), &deviceLocations)
			if nil != err {
				return err
			}
		}

		return nil
	})
}

func (self *User) ExportMeasurementStatsByLocation() ([]*MeasurementLocations, error) {
	var measurementLocations []*MeasurementLocations

	return measurementLocations, self.db.Exec(func(conn *sql.DB) error {

		rows, err := conn.Query(`
			WITH measurements_locations AS (
			        SELECT
			            measurements.key,
			            COUNT(DISTINCT(measurements.sensor_id)) AS sensors,
			            COUNT(measurements.*),
			            MIN(measurements.value),
			            MAX(measurements.value),
			            STDDEV(measurements.value),
			            AVG(measurements.value),
						to_char(MIN(measurements.created_at), 'YYYY-MM-DD"T"HH:MI:SS"Z"') AS first_timestamp,
						to_char(MAX(measurements.created_at), 'YYYY-MM-DD"T"HH:MI:SS"Z"') AS lastest_timestamp,
			            locations.id AS location_id,
			            locations.name AS location_name,
			            ST_AsGeoJSON(locations.geom)::JSONB AS geometry
			        FROM measurements
			            INNER JOIN locations
			                ON locations.id = measurements.location_id
			        INNER JOIN sensors
			            ON sensors.id = measurements.sensor_id
			            AND sensors.is_deleted = false
			        INNER JOIN devices
			            ON devices.id = sensors.device_id
			            ANd devices.username = $1
			            AND devices.is_deleted = false
			        GROUP BY
			            locations.id, measurements.key
			    )

			SELECT json_agg(c)
			FROM (
			    SELECT
			        location_id,
			        location_name,
			        geometry,
			        json_agg(
			            json_build_object(
			                'key',
			                key,
			                'sensors',
			                sensors,
							'count',
							count,
			                'min',
			                min,
			                'max',
			                max,
			                'stddev',
			                stddev,
			                'mean',
			                avg,
			                'first_timestamp',
			                first_timestamp,
			                'lastest_timestamp',
			                lastest_timestamp
			            )
			        ) AS scanners
			    FROM
			        measurements_locations
			    GROUP BY
			        location_id,
			        location_name,
			        geometry
			) c; `, self.Username)

		if nil != err {
			return err
		}

		for rows.Next() {
			var temp string
			rows.Scan(&temp)
			err = json.Unmarshal([]byte(temp), &measurementLocations)
			if nil != err {
				return err
			}
		}

		return nil
	})
}
