package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/karlseguin/ccache"
	"github.com/sjsafranek/find5/database"
	"github.com/sjsafranek/ligneous"
)

var (
	logger = ligneous.AddLogger("api", "trace", "./log/find5")
)

func New(dbConnStr string, redisAddr string) *Api {
	return &Api{
		db: database.New(dbConnStr),
		redis: &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", redisAddr) },
		},
		cache: ccache.Layered(ccache.Configure()),
	}
}

type Api struct {
	db    *database.Database
	redis *redis.Pool
	cache *ccache.LayeredCache
}

func (self *Api) fetchUser(request *Request) (*database.User, error) {
	var user *database.User
	var err error
	if "" != request.Apikey {
		user, err = self.GetUserByApikey(request.Apikey)
	} else if "" != request.Username {
		user, err = self.GetUserByUsername(request.Username)
	} else {
		err = errors.New("Missing parameters")
	}
	return user, err
}

// CreateUser
func (self *Api) CreateUser(email, username, password string) (*database.User, error) {
	user, err := self.db.CreateUser(email, username, password)
	if nil == err {
		// cache apikey user pair
		self.cache.Set("user", user.Apikey, user, 5*time.Minute)
	}
	return user, err
}

// GetUserByUserName
func (self *Api) GetUserByUsername(username string) (*database.User, error) {
	return self.db.GetUserFromUsername(username)
}

// GetUserByApikey fetches user via apikey. This method uses an inmemory LRU cache to
// decrease the number of database transactions.
func (self *Api) GetUserByApikey(apikey string) (*database.User, error) {
	// check cache for apikey user pair
	item := self.cache.Get("user", apikey)
	if nil != item {
		return item.Value().(*database.User), nil
	}

	user, err := self.db.GetUserFromApikey(apikey)
	if nil == err {
		// cache apikey user pair
		self.cache.Set("user", apikey, user, 5*time.Minute)
	}
	return user, err
}

// fetch device
func (self *Api) fetchDevice(request *Request) (*database.Device, error) {
	var device *database.Device

	item := self.cache.Get("device_id", request.DeviceId)
	if nil != item {
		device = item.Value().(*database.Device)
	} else {

		user, err := self.fetchUser(request)
		if nil != err {
			return device, err
		}

		device, err = user.GetDeviceById(request.DeviceId)
		if nil != err {
			return device, err
		}

		// cache device device_id pair
		self.cache.Set("device", request.DeviceId, device, 5*time.Minute)
	}

	return device, nil
}

//
func (self *Api) fetchSensor(request *Request) (*database.Sensor, error) {
	var sensor *database.Sensor
	device, err := self.fetchDevice(request)
	if nil != err {
		return sensor, nil
	}
	return device.GetSensorById(request.SensorId)
}

// RecordMeasurements
func (self *Api) RecordMeasurements(request *Request) error {
	device, err := self.fetchDevice(request)
	if nil != err {
		return err
	}

	for sensor_id := range request.Data {
		sensor, err := device.GetSensorById(sensor_id)
		if nil != err {
			return err
		}
		if "" != request.LocationId {
			sensor.RecordMeasurementsAtLocation(request.LocationId, request.Data[sensor_id])
		} else {
			sensor.RecordMeasurements(request.Data[sensor_id])
		}
	}

	return nil
}

func (self *Api) Do(request *Request) (*Response, error) {
	var response Response

	response.Status = "ok"

	err := func() error {
		switch request.Method {

		case "ping":
			// {"method":"ping"}
			response.Message = "pong"
			return nil

		case "create_user":
			// {"method":"create_user","username": "admin_user" "email":"admin@email.com","password":"1234"}
			if "" == request.Email || "" == request.Username {
				return errors.New("Missing parameters")
			}

			user, err := self.CreateUser(request.Email, request.Username, request.Password)
			if nil != err {
				return err
			}

			response.Data.User = user
			return nil

		case "get_users":
			// {"method":"get_users"}
			users, err := self.db.GetUsers()
			if nil != err {
				return err
			}
			response.Data.Users = users
			return nil

		case "get_user":
			// {"method":"get_user","username":"admin_user"}
			// {"method":"get_user","apikey":"<apikey>"}
			user, err := self.fetchUser(request)
			if nil != err {
				return err
			}

			response.Data.User = user
			return nil

		case "delete_user":
			// {"method":"delete_user","username":"admin_user"}
			// {"method":"delete_user","apikey":"<apikey>"}
			user, err := self.fetchUser(request)
			if nil != err {
				return err
			}

			err = user.Delete()
			if nil != err {
				return err
			}

			return nil

		case "set_password":
			// {"method":"set_password","username":"admin_user","password":"1234"}
			// {"method":"set_password","apikey":"<apikey>","password":"1234"}
			user, err := self.fetchUser(request)
			if nil != err {
				return err
			}

			// TODO require to not be empty string...
			err = user.SetPassword(request.Password)
			if nil != err {
				return err
			}

			return nil

		case "create_device":
			// {"method":"create_device","username":"admin_user","name":"laptop","type":"computer"}
			// {"method":"create_device","apikey":"<apikey>","name":"laptop","type":"computer"}
			user, err := self.fetchUser(request)
			if nil != err {
				return err
			}

			device, err := user.CreateDevice(request.Name, request.Type)
			if nil != err {
				return err
			}

			// cache device
			self.cache.Set("device", device.Id, device, 5*time.Minute)

			response.Data.Device = device
			return nil

		case "get_devices":
			// {"method":"get_devices","username":"admin_user"}
			// {"method":"get_devices","apikey":"<apikey>"}
			user, err := self.fetchUser(request)
			if nil != err {
				return err
			}

			devices, err := user.GetDevices()
			if nil != err {
				return err
			}

			response.Data.Devices = devices
			return nil

		case "get_device":
			// {"method":"get_device","username":"admin_user","device_id":"<uuid>"}
			// {"method":"get_device","apikey":"<apikey>","device_id":"<uuid>"}
			device, err := self.fetchDevice(request)
			if nil != err {
				return err
			}
			fmt.Println(device)
			response.Data.Device = device
			return nil

		case "create_sensor":
			// {"method":"create_sensor","username":"admin_user","device_id":"<uuid>","name":"laptop","type":"computer"}
			// {"method":"create_sensor","apikey":"<apikey>","device_id":"<uuid>","name":"laptop","type":"computer"}
			device, err := self.fetchDevice(request)
			if nil != err {
				return err
			}

			err = device.CreateSensor(request.Name, request.Type)
			if nil != err {
				return err
			}

			return nil

		case "get_sensors":
			// {"method":"get_sensors","username":"admin_user","device_id":"<uuid>"}
			// {"method":"get_sensors","apikey":"<apikey>","device_id":"<uuid>"}
			device, err := self.fetchDevice(request)
			if nil != err {
				return err
			}

			sensors, err := device.GetSensors()
			if nil != err {
				return err
			}

			response.Data.Sensors = sensors
			return nil

		case "get_sensor":
			// {"method":"get_sensor","username":"admin_user","sensor_id":"<uuid>"}
			// {"method":"get_sensor","apikey":"<apikey>","sensor_id":"<uuid>"}
			sensor, err := self.fetchSensor(request)
			if nil != err {
				return err
			}

			response.Data.Sensor = sensor
			return nil

		case "create_location":
			// {"method":"create_location","username":"admin_user","name":"MyHouse","longitude":0.0,"latitude":0.0}
			// {"method":"create_location","apikey":"<apikey>","name":"MyHouse","longitude":0.0,"latitude":0.0}
			user, err := self.fetchUser(request)
			if nil != err {
				return err
			}

			err = user.CreateLocation(request.Name, request.Longitude, request.Latitude)
			if nil != err {
				return err
			}

			return nil

		case "get_locations":
			// {"method":"get_locations","username":"admin_user"}
			// {"method":"get_locations","apikey":"<apikey>"}
			user, err := self.fetchUser(request)
			if nil != err {
				return err
			}

			locations, err := user.GetLocations()
			if nil != err {
				return err
			}

			response.Data.Locations = locations
			return nil

		case "record_measurements":
			// {"method":"record_measurements","username":"admin","device_id":"dada27ee-f57b-e9c0-4ac0-1b2eda8af6fb","data":{"5c434f17-3095-7f74-7688-9de7f7853c2d":{"thing1":1,"thing2":2, "thing3":3}}}
			return self.RecordMeasurements(request)

		default:
			return errors.New("Method not found")

		}
	}()

	if nil != err {
		response.SetError(err)
	}

	return &response, err
}
