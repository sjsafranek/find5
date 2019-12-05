package api

import (
	"encoding/json"
	"errors"
	// "strings"
	"time"

	// "github.com/garyburd/redigo/redis"
	"github.com/karlseguin/ccache"
	"github.com/sjsafranek/find5/findapi/lib/ai"
	"github.com/sjsafranek/find5/findapi/lib/ai/models"
	"github.com/sjsafranek/find5/findapi/lib/database"
	"github.com/sjsafranek/ligneous"
)

var (
	logger = ligneous.AddLogger("api", "trace", "")
)

func SetLoggingDirectory(directory string) {
	logger = ligneous.AddLogger("api", "trace", directory)
	database.SetLoggingDirectory(directory)
	ai.SetLoggingDirectory(directory)
}

func New(dbConnStr, aiConnStr, redisAddr string) *Api {
	return &Api{
		db: database.New(dbConnStr),
		// redis: &redis.Pool{
		// 	MaxIdle:     3,
		// 	IdleTimeout: 240 * time.Second,
		// 	Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", redisAddr) },
		// },
		cache: ccache.Layered(ccache.Configure()),
		ai:    ai.New(aiConnStr, redisAddr),
	}
}

type Api struct {
	db *database.Database
	// redis          *redis.Pool
	cache          *ccache.LayeredCache
	ai             *ai.AI
	eventListeners map[string][]func(string, string, float64)
}

func (self *Api) RegisterEventListener(username string, clbk func(string, string, float64)) {
	if nil == self.eventListeners {
		self.eventListeners = make(map[string][]func(string, string, float64))
	}
	self.eventListeners[username] = append(self.eventListeners[username], clbk)
}

func (self *Api) fireEvent(username, device_id, location_id string, probability float64) {
	logger.Info(username, device_id, location_id, probability)
	if clbks, ok := self.eventListeners[username]; ok {
		for i := 0; i < len(clbks); i++ {
			clbks[i](device_id, location_id, probability)
		}
	}
}

func (self *Api) fetchUser(request *Request, clbk func(*database.User) error) error {
	var user *database.User
	var err error
	if "" != request.Apikey {
		user, err = self.getUserByApikey(request.Apikey)
	} else if "" != request.Username {
		user, err = self.getUserByUsername(request.Username)
	} else {
		err = errors.New("Missing parameters")
	}
	if nil != err {
		return err
	}
	return clbk(user)
}

// CreateUser
func (self *Api) createUser(email, username, password string) (*database.User, error) {
	user, err := self.db.CreateUser(email, username, password)
	if nil == err {
		// cache apikey user pair
		self.cache.Set("user", user.Apikey, user, 5*time.Minute)
	}
	return user, err
}

// GetUserByUserName
func (self *Api) getUserByUsername(username string) (*database.User, error) {
	return self.db.GetUserFromUsername(username)
}

// GetUserByApikey fetches user via apikey. This method uses an inmemory LRU cache to
// decrease the number of database transactions.
func (self *Api) getUserByApikey(apikey string) (*database.User, error) {
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
func (self *Api) fetchDevice(request *Request, clbk func(*database.Device) error) error {
	var device *database.Device

	item := self.cache.Get("device_id", request.DeviceId)
	if nil != item {
		device = item.Value().(*database.Device)
	} else {
		err := self.fetchUser(request, func(user *database.User) error {
			var err error
			device, err = user.GetDeviceById(request.DeviceId)
			return err
		})

		if nil != err {
			return err
		}

		self.cache.Set("device", request.DeviceId, device, 5*time.Minute)
	}

	return clbk(device)
}

//
func (self *Api) fetchSensor(request *Request, clbk func(*database.Sensor) error) error {
	return self.fetchDevice(request, func(device *database.Device) error {
		sensor, err := device.GetSensorById(request.SensorId)
		if nil != err {
			return err
		}
		return clbk(sensor)
	})
}

func (self *Api) AnalyzeData(request *Request) error {
	return self.fetchDevice(request, func(device *database.Device) error {

		sd := models.SensorData{
			Family:    device.GetUser().Username,
			Device:    request.DeviceId,
			Location:  request.LocationId,
			Timestamp: time.Now().Unix(),
			Sensors:   make(map[string]map[string]interface{}),
		}

		for i := range request.Data {
			if _, ok := sd.Sensors[i]; !ok {
				sd.Sensors[i] = make(map[string]interface{})
			}
			for j := range request.Data[i] {
				sd.Sensors[i][j] = request.Data[i][j]
			}
		}

		aidata, err := self.ai.Analyze(sd, device.GetUser().Username)
		// add location predictions
		if nil == err {
			go func() {
				logger.Info(aidata.Guesses)
				for i := range aidata.Guesses {
					aidata.Guesses[i].Probability = float64(int64(float64(aidata.Guesses[i].Probability)*100)) / 100
					device.SetLocation(aidata.Guesses[i].Location, aidata.Guesses[i].Probability*100)
					self.fireEvent(device.GetUser().Username, device.Id, aidata.Guesses[i].Location, aidata.Guesses[i].Probability*100)
				}
			}()
		}
		//.end

		return err
	})
}

// Calibrate
func (self *Api) calibrate(request *Request) error {
	return self.fetchUser(request, func(user *database.User) error {
		measurements, err := user.ExportMeasurements()
		if nil != err {
			return err
		}
		return self.ai.Calibrate(measurements, user.Username, true)
	})
}

// RecordMeasurements
func (self *Api) importMeasurements(request *Request) error {
	return self.fetchDevice(request, func(device *database.Device) error {
		if !device.IsActive {
			return errors.New("device is deactivated")
		}

		// TESTING
		go func() {
			err := self.AnalyzeData(request)
			if nil != err {
				// // HACK
				// //  - determine if calibration is needed
				// if strings.Contains(err.Error(), "could not find") {
				// 	logger.Warn("Data not calibrated")
				// 	err = self.calibrate(request)
				// 	if nil != err {
				// 		logger.Error(err)
				// 	}
				// }
				logger.Error(err)
			}
		}()
		//.end

		for sensor_id := range request.Data {
			sensor, err := device.GetSensorById(sensor_id)
			if nil != err {
				return err
			}
			if !sensor.IsActive {
				return errors.New("sensor is deactivated")
			}

			if "" != request.LocationId {
				err = sensor.ImportMeasurementsAtLocation(request.LocationId, request.Data[sensor_id])
			} else {
				err = sensor.ImportMeasurements(request.Data[sensor_id])
			}

			if nil != err {
				return nil
			}

		}
		return nil
	})

}

func (self *Api) DoJSON(jdata string) (*Response, error) {
	var request Request
	err := json.Unmarshal([]byte(jdata), &request)
	if nil != err {
		response := &Response{Status: "err"}
		response.SetError(err)
		return response, err
	}
	return self.Do(&request)
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
			if "" == request.Username {
				return errors.New("missing parameters")
			}

			user, err := self.createUser(request.Email, request.Username, request.Password)
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
			return self.fetchUser(request, func(user *database.User) error {
				response.Data.User = user
				return nil
			})

		case "delete_user":
			// {"method":"delete_user","username":"admin_user"}
			// {"method":"delete_user","apikey":"<apikey>"}
			return self.fetchUser(request, func(user *database.User) error {
				self.cache.Delete("user", user.Apikey)
				return user.Delete()
			})

		case "activate_user":
			// {"method":"activate_user","username":"admin_user"}
			// {"method":"activate_user","apikey":"<apikey>"}
			return self.fetchUser(request, func(user *database.User) error {
				self.cache.Delete("user", user.Apikey)
				return user.Activate()
			})

		case "deactivate_user":
			// {"method":"deactivate_user","username":"admin_user"}
			// {"method":"deactivate_user","apikey":"<apikey>"}
			return self.fetchUser(request, func(user *database.User) error {
				self.cache.Delete("user", user.Apikey)
				return user.Deactivate()
			})

		case "set_password":
			// {"method":"set_password","username":"admin_user","password":"1234"}
			// {"method":"set_password","apikey":"<apikey>","password":"1234"}
			return self.fetchUser(request, func(user *database.User) error {
				return user.SetPassword(request.Password)
			})

		case "create_device":
			// {"method":"create_device","username":"admin_user","name":"laptop","type":"computer"}
			// {"method":"create_device","apikey":"<apikey>","name":"laptop","type":"computer"}
			return self.fetchUser(request, func(user *database.User) error {
				device, err := user.CreateDevice(request.Name, request.Type)
				if nil != err {
					return err
				}

				// cache device
				self.cache.Set("device", device.Id, device, 5*time.Minute)
				response.Data.Device = device
				return nil
			})

		case "get_devices":
			// {"method":"get_devices","username":"admin_user"}
			// {"method":"get_devices","apikey":"<apikey>"}
			return self.fetchUser(request, func(user *database.User) error {
				devices, err := user.GetDevices()
				if nil != err {
					return err
				}
				response.Data.Devices = devices
				return nil
			})

		case "get_device":
			// {"method":"get_device","username":"admin_user","device_id":"<uuid>"}
			// {"method":"get_device","apikey":"<apikey>","device_id":"<uuid>"}
			return self.fetchDevice(request, func(device *database.Device) error {
				response.Data.Device = device
				return nil
			})

		case "delete_device":
			// {"method":"delete_device","username":"admin_user","device_id":"<uuid>"}
			// {"method":"delete_device","apikey":"<apikey>","device_id":"<uuid>"}
			return self.fetchDevice(request, func(device *database.Device) error {
				self.cache.Delete("device", request.DeviceId)
				return device.Delete()
			})

		case "activate_device":
			// {"method":"activate_device","username":"admin_user","device_id":"<uuid>"}
			// {"method":"activate_device","apikey":"<apikey>","device_id":"<uuid>"}
			return self.fetchDevice(request, func(device *database.Device) error {
				err := device.Activate()
				self.cache.Replace("device", request.DeviceId, device)
				return err
			})

		case "deactivate_device":
			// {"method":"deactivate_device","username":"admin_user","device_id":"<uuid>"}
			// {"method":"deactivate_device","apikey":"<apikey>","device_id":"<uuid>"}
			return self.fetchDevice(request, func(device *database.Device) error {
				err := device.Deactivate()
				self.cache.Replace("device", request.DeviceId, device)
				return err
			})

		case "create_sensor":
			// {"method":"create_sensor","username":"admin_user","device_id":"<uuid>","name":"laptop","type":"computer"}
			// {"method":"create_sensor","apikey":"<apikey>","device_id":"<uuid>","name":"laptop","type":"computer"}
			return self.fetchDevice(request, func(device *database.Device) error {
				if !device.IsActive {
					return errors.New("device is deactivated")
				}
				return device.CreateSensor(request.Name, request.Type)
			})

		case "get_sensors":
			// {"method":"get_sensors","username":"admin_user","device_id":"<uuid>"}
			// {"method":"get_sensors","apikey":"<apikey>","device_id":"<uuid>"}
			return self.fetchDevice(request, func(device *database.Device) error {
				sensors, err := device.GetSensors()
				if nil != err {
					return err
				}
				response.Data.Sensors = sensors
				return nil
			})

		case "get_sensor":
			// {"method":"get_sensor","username":"admin_user","sensor_id":"<uuid>"}
			// {"method":"get_sensor","apikey":"<apikey>","sensor_id":"<uuid>"}
			return self.fetchSensor(request, func(sensor *database.Sensor) error {
				response.Data.Sensor = sensor
				return nil
			})

		case "delete_sensor":
			// {"method":"delete_sensor","username":"admin_user","sensor_id":"<uuid>"}
			// {"method":"delete_sensor","apikey":"<apikey>","sensor_id":"<uuid>"}
			return self.fetchSensor(request, func(sensor *database.Sensor) error {
				self.cache.Delete("sensor", request.SensorId)
				return sensor.Delete()
			})

		case "activate_sensor":
			// {"method":"activate_sensor","username":"admin_user","sensor_id":"<uuid>"}
			// {"method":"activate_sensor","apikey":"<apikey>","sensor_id":"<uuid>"}
			return self.fetchSensor(request, func(sensor *database.Sensor) error {
				err := sensor.Activate()
				self.cache.Replace("sensor", request.SensorId, sensor)
				return err
			})

		case "deactivate_sensor":
			// {"method":"deactivate_sensor","username":"admin_user","sensor_id":"<uuid>"}
			// {"method":"deactivate_sensor","apikey":"<apikey>","sensor_id":"<uuid>"}
			return self.fetchSensor(request, func(sensor *database.Sensor) error {
				err := sensor.Deactivate()
				self.cache.Replace("sensor", request.SensorId, sensor)
				return err
			})

		case "create_location":
			// {"method":"create_location","username":"admin_user","name":"MyHouse","longitude":0.0,"latitude":0.0}
			// {"method":"create_location","apikey":"<apikey>","name":"MyHouse","longitude":0.0,"latitude":0.0}
			return self.fetchUser(request, func(user *database.User) error {
				return user.CreateLocation(request.Name, request.Longitude, request.Latitude)
			})

		case "get_locations":
			// {"method":"get_locations","username":"admin_user"}
			// {"method":"get_locations","apikey":"<apikey>"}
			return self.fetchUser(request, func(user *database.User) error {
				locations, err := user.GetLocations()
				if nil != err {
					return err
				}

				response.Data.Locations = locations
				return nil
			})

		case "delete_location":
			// TODO
			return nil

		case "import_measurements":
			// {"method":"record_measurements","username":"admin","device_id":"dada27ee-f57b-e9c0-4ac0-1b2eda8af6fb","data":{"5c434f17-3095-7f74-7688-9de7f7853c2d":{"thing1":1,"thing2":2, "thing3":3}}}
			return self.importMeasurements(request)

		case "export_measurements":
			// {"method":"export_measurements","username":"admin"}
			return self.fetchUser(request, func(user *database.User) error {
				measurements, err := user.ExportMeasurements()
				if nil != err {
					return err
				}
				response.Data.Measurements = measurements
				return nil
			})

		case "export_devices_by_location":
			// {"method":"export_devices_by_location","username":"admin"}
			return self.fetchUser(request, func(user *database.User) error {
				deviceLocations, err := user.ExportDevicesByLocation()
				if nil != err {
					return err
				}
				response.Data.DeviceLocations = deviceLocations
				return nil
			})

		case "export_measurement_stats_by_location":
			// {"method":"export_measurement_stats_by_location","username":"admin"}
			return self.fetchUser(request, func(user *database.User) error {
				measurementLocations, err := user.ExportMeasurementStatsByLocation()
				if nil != err {
					return err
				}
				response.Data.MeasurementLocations = measurementLocations
				return nil
			})

		case "calibrate":
			// {"method":"calibrate","username":"admin"}
			// {"apikey": "a0a1695e8cd13322f1acb312b40cddb6", "method": "calibrate"}
			return self.calibrate(request)
			// return self.fetchUser(request, func(user *database.User) error {
			// 	measurements, err := user.ExportMeasurements()
			// 	if nil != err {
			// 		return err
			// 	}
			// 	// TODO
			// 	return self.ai.Calibrate(measurements, user.Username, true)
			// })

		default:
			return errors.New("method not found")

		}
	}()

	if nil != err {
		response.SetError(err)
	}

	return &response, err
}
