package main

import (
	"flag"
	"fmt"

	"github.com/sjsafranek/find5/database"
	"github.com/sjsafranek/ligneous"
)

const (
	DEFAULT_PORT int    = 5555
	PROJECT      string = "find5"
	VERSION      string = "0.0.1"
)

const (
	DEFAULT_DATABASE_ENGINE   string = "postgres"
	DEFAULT_DATABASE_DATABASE string = "finddb"
	DEFAULT_DATABASE_PASSWORD string = "dev"
	DEFAULT_DATABASE_USERNAME string = "finduser"
	DEFAULT_DATABASE_HOST     string = "localhost"
	DEFAULT_DATABASE_PORT     int64  = 5432
)

var (
	DATABASE_ENGINE   string = DEFAULT_DATABASE_ENGINE
	DATABASE_DATABASE string = DEFAULT_DATABASE_DATABASE
	DATABASE_PASSWORD string = DEFAULT_DATABASE_PASSWORD
	DATABASE_USERNAME string = DEFAULT_DATABASE_USERNAME
	DATABASE_HOST     string = DEFAULT_DATABASE_HOST
	DATABASE_PORT     int64  = DEFAULT_DATABASE_PORT
)

var (
	PORT   int = DEFAULT_PORT
	logger     = ligneous.AddLogger("server", "trace", "./log/find5")
)

func init() {
	var print_version bool

	// flag.StringVar(&ACTION, "action", DEFAULT_ACTION, "Action")
	// flag.StringVar(&CONFIG_FILE, "c", DEFAULT_CONFIG_FILE, "Config file")
	// flag.BoolVar(&DEBUG, "debug", false, "debug mode")

	flag.IntVar(&PORT, "port", DEFAULT_PORT, "Server port")
	flag.StringVar(&DATABASE_HOST, "dbhost", DEFAULT_DATABASE_HOST, "database host")
	flag.StringVar(&DATABASE_DATABASE, "dbname", DEFAULT_DATABASE_DATABASE, "database name")
	flag.StringVar(&DATABASE_PASSWORD, "dbpass", DEFAULT_DATABASE_PASSWORD, "database password")
	flag.StringVar(&DATABASE_USERNAME, "dbuser", DEFAULT_DATABASE_USERNAME, "database username")
	flag.Int64Var(&DATABASE_PORT, "dbport", DEFAULT_DATABASE_PORT, "Database port")

	flag.BoolVar(&print_version, "V", false, "Print version and exit")

	flag.Parse()

}

func main() {

	dbConnectionString := fmt.Sprintf("%v://%v:%v@%v:%v/%v?sslmode=disable", DATABASE_ENGINE, DATABASE_USERNAME, DATABASE_PASSWORD, DATABASE_HOST, DATABASE_PORT, DATABASE_DATABASE)
	db := database.New(dbConnectionString)

	// variables
	var user *database.User
	var err error

	// setup
	// user, err = db.CreateUser("admin@find.com", "admin_user", "dev")
	// if nil != err {
	// 	panic(err)
	// }
	//
	// err = user.CreateDevice("laptop", "computer")
	// if nil != err {
	// 	panic(err)
	// }
	//
	// err = user.CreateLocation("office_desk", 0, 0)
	// if nil != err {
	// 	panic(err)
	// }
	//
	// devices, err := user.GetDevices()
	// if nil != err {
	// 	panic(err)
	// }
	//
	// err = devices[0].CreateSensor("wifi_card", "mac_addresses")
	// if nil != err {
	// 	panic(err)
	// }

	// test
	user, err = db.GetUserFromUsername("admin_user")
	if nil != err {
		panic(err)
	}

	device, err := user.GetDeviceByName("laptop")
	if nil != err {
		panic(err)
	}

	fc, err := user.GetLocations()
	if nil != err {
		panic(err)
	}
	location_id := fc.Features[0].Properties["id"].(string)

	sensor, err := device.GetSensorByName("wifi_card")
	if nil != err {
		panic(err)
	}

	err = sensor.RecordMeasurementAtLocation(location_id, "test", 23)
	if nil != err {
		panic(err)
	}

	err = sensor.RecordMeasurement("test", 123)
	if nil != err {
		panic(err)
	}

	data := make(map[string]float64)
	data["thing1"] = 0.0
	data["thing2"] = 4.2
	data["thing3"] = 66.6
	err = sensor.RecordMeasurementsAtLocation(location_id, data)
	if nil != err {
		panic(err)
	}

	err = sensor.RecordMeasurements(data)
	if nil != err {
		panic(err)
	}

}
