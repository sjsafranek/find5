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

type User interface {
	SetPassword(string) error
}

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

func (self *Api) Execute(query *Query) (string, error) {
	switch query.Method {

	case "ping":
		return `{"status":"ok","message":"pong"}`, nil

	case "create_user":
		if "" == query.Email || "" == query.Username {
			err := errors.New("Missing parameters")
			return fmt.Sprintf(`{"status":"error","error":"%v"}`, err.Error()), err
		}

		user, err := self.CreateUser(query.Email, query.Username, query.Password)
		if nil != err {
			return fmt.Sprintf(`{"status":"error","error":"%v"}`, err.Error()), err
		}

		data, err := user.Marshal()
		if nil != err {
			return fmt.Sprintf(`{"status":"error","error":"%v"}`, err.Error()), err
		}
		return fmt.Sprintf(`{"status":"ok","data":{"user":%v}}`, data), nil

	case "get_users":
		return `{"status":"ok","message":"TODO"}`, nil

	case "get_user":
		user, err := self.fetchUser(query)
		if nil != err {
			return fmt.Sprintf(`{"status":"error","error":"%v"}`, err.Error()), err
		}

		data, err := user.Marshal()
		if nil != err {
			return fmt.Sprintf(`{"status":"error","error":"%v"}`, err.Error()), err
		}

		return fmt.Sprintf(`{"status":"ok","data":{"user":%v}}`, data), nil

	case "delete_user":
		user, err := self.fetchUser(query)
		if nil != err {
			return fmt.Sprintf(`{"status":"error","error":"%v"}`, err.Error()), err
		}

		err = user.Delete()
		if nil != err {
			return fmt.Sprintf(`{"status":"error","error":"%v"}`, err.Error()), err
		}

		return `{"status":"ok"}`, nil

	case "set_password":
		user, err := self.fetchUser(query)
		if nil != err {
			return fmt.Sprintf(`{"status":"error","error":"%v"}`, err.Error()), err
		}

		// TODO require to not be empty string...
		err = user.SetPassword(query.Password)
		if nil != err {
			return fmt.Sprintf(`{"status":"error","error":"%v"}`, err.Error()), err
		}
		return `{"status":"ok"}`, nil

	case "create_device":
		user, err := self.fetchUser(query)
		if nil != err {
			return fmt.Sprintf(`{"status":"error","error":"%v"}`, err.Error()), err
		}

		device, err := user.CreateDevice(query.Name, query.Type)
		if nil != err {
			return fmt.Sprintf(`{"status":"error","error":"%v"}`, err.Error()), err
		}

		data, err := device.Marshal()
		if nil != err {
			return fmt.Sprintf(`{"status":"error","error":"%v"}`, err.Error()), err
		}

		return fmt.Sprintf(`{"status":"ok","data":{"device":%v}}`, data), nil

	case "get_devices":
		return `{"status":"ok","message":"TODO"}`, nil

	case "get_device":
		return `{"status":"ok","message":"TODO"}`, nil

	case "create_sensor":
		return `{"status":"ok","message":"TODO"}`, nil

	case "get_sensors":
		return `{"status":"ok","message":"TODO"}`, nil

	case "get_sensor":
		return `{"status":"ok","message":"TODO"}`, nil

	case "get_locations":
		return `{"status":"ok","message":"TODO"}`, nil

	case "create_location":
		return `{"status":"ok","message":"TODO"}`, nil

	case "record_data":
		return `{"status":"ok","message":"TODO"}`, nil

	default:
		return `{"status":"error","error":"method not found"}`, nil

	}
}

func (self *Api) fetchUser(query *Query) (*database.User, error) {
	var user *database.User
	var err error
	if "" != query.Apikey {
		user, err = self.GetUserByApikey(query.Apikey)
	} else if "" != query.Username {
		user, err = self.GetUserByUsername(query.Username)
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
func (self *Api) fetchDeviceById(apikey, device_id string) (*database.Device, error) {
	var device *database.Device

	item := self.cache.Get("device_id", device_id)
	if nil != item {
		device = item.Value().(*database.Device)
	} else {
		user, err := self.GetUserByApikey(apikey)
		if nil != err {
			return device, err
		}

		device, err := user.GetDeviceById(device_id)
		if nil != err {
			return device, err
		}

		// cache device device_id pair
		self.cache.Set("device", device_id, device, 5*time.Minute)
	}

	return device, nil
}

// RecordMeasurements
func (self *Api) RecordMeasurements(apikey, device_id string, data map[string]map[string]float64) error {
	device, err := self.fetchDeviceById(apikey, device_id)
	if nil != err {
		return err
	}

	for sensor_id := range data {
		sensor, err := device.GetSensorById(sensor_id)
		if nil != err {
			return err
		}
		sensor.RecordMeasurements(data[sensor_id])
	}

	return nil
}

// RecordMeasurementsAtLocation
func (self *Api) RecordMeasurementsAtLocation(apikey, device_id, location_id string, data map[string]map[string]float64) error {
	device, err := self.fetchDeviceById(apikey, device_id)
	if nil != err {
		return err
	}

	for sensor_id := range data {
		sensor, err := device.GetSensorById(sensor_id)
		if nil != err {
			return err
		}
		sensor.RecordMeasurementsAtLocation(location_id, data[sensor_id])
	}

	return nil
}
