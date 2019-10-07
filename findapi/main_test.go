package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/sjsafranek/find5/findapi/lib/api"
	"github.com/sjsafranek/find5/findapi/lib/database"
)

/*
DELETE from users where username like 'test__%';




{"method": "create_sensor", "type": "wifi", "device_id": "a863c68f-9010-8afe-a12d-738ae312b7ad", "name": "wifi_card", "apikey": "c2d7589a472941efb29577833dd6a5fb"}
{"method": "get_sensors", "apikey": "c2d7589a472941efb29577833dd6a5fb", "device_id": "a863c68f-9010-8afe-a12d-738ae312b7ad"}
a863c68f-9010-8afe-a12d-738ae312b7ad c113b56e-f03c-dfc5-6ead-c458373b3f5d
{"method": "create_location", "apikey": "c2d7589a472941efb29577833dd6a5fb", "longitude": 0.0, "latitude": 0.0, "name": "zakhome floor 2 bedroom"}
{"method": "create_location", "apikey": "c2d7589a472941efb29577833dd6a5fb", "longitude": 0.0, "latitude": 0.0, "name": "zakhome floor 1 kitchen"}
{"method": "create_location", "apikey": "c2d7589a472941efb29577833dd6a5fb", "longitude": 0.0, "latitude": 0.0, "name": "zakhome floor 2 office"}
{"method": "get_locations", "apikey": "c2d7589a472941efb29577833dd6a5fb"}



{"method": "import_measurements", "location_id": "01e25343-80a1-c736-dc6e-16914ccf92dc", "device_id": "a863c68f-9010-8afe-a12d-738ae312b7ad", "apikey": "c2d7589a472941efb29577833dd6a5fb", "data": {"c113b56e-f03c-dfc5-6ead-c458373b3f5d": {"80:37:73:ba:f7:dc": -58, "00:1a:1e:46:cd:11": -84, "f8:35:dd:0a:da:be": -84, "a0:63:91:2b:9e:64": -68, "30:46:9a:a0:28:c4": -76, "e8:ed:05:55:21:10": -83, "e0:46:9a:6d:02:ea": -84, "00:23:69:d4:47:9f": -75, "80:37:73:ba:f7:d8": -45, "b8:3e:59:78:35:99": -88, "20:aa:4b:b8:31:c8": -82, "d4:05:98:57:b3:10": -75, "00:1a:1e:46:cd:10": -82, "2c:b0:5d:36:e3:b8": -81, "ec:1a:59:4a:9c:ed": -83, "70:73:cb:bd:9f:b5": -70, "b4:75:0e:03:cd:69": -84, "a0:63:91:2b:9e:65": -61}}}

{"method": "import_measurements", "location_id": "01e25343-80a1-c736-dc6e-16914ccf92dc", "device_id": "a863c68f-9010-8afe-a12d-738ae312b7ad", "apikey": "c2d7589a472941efb29577833dd6a5fb", "data": {"c113b56e-f03c-dfc5-6ead-c458373b3f5d": {"80:37:73:ba:f7:dc": -56, "00:1a:1e:46:cd:11": -81, "f8:35:dd:0a:da:be": -84, "34:68:95:f8:25:fd": -74, "b8:3e:59:78:35:99": -86, "00:23:69:d4:47:9f": -75, "80:37:73:ba:f7:d8": -43, "20:aa:4b:b8:31:c8": -83, "00:1a:1e:46:cd:10": -81, "2c:b0:5d:36:e3:b8": -81, "e0:46:9a:6d:02:ea": -84, "b4:75:0e:03:cd:66": -83, "a0:63:91:2b:9e:64": -66, "30:46:9a:a0:28:c4": -84, "e8:ed:05:55:21:10": -83, "4c:60:de:fe:e5:24": -85, "b4:75:0e:03:cd:69": -84, "d4:05:98:57:b3:10": -79, "ec:1a:59:4a:9c:ed": -83, "70:73:cb:bd:9f:b5": -69, "00:ac:e0:b8:ea:a0": -81, "a0:63:91:2b:9e:65": -60}}}

{"method": "import_measurements", "location_id": "01e25343-80a1-c736-dc6e-16914ccf92dc", "device_id": "a863c68f-9010-8afe-a12d-738ae312b7ad", "apikey": "c2d7589a472941efb29577833dd6a5fb", "data": {"c113b56e-f03c-dfc5-6ead-c458373b3f5d": {"80:37:73:ba:f7:dc": -45, "00:1a:1e:46:cd:11": -76, "f8:35:dd:0a:da:be": -84, "34:68:95:f8:25:fd": -74, "b8:3e:59:78:35:99": -86, "00:23:69:d4:47:9f": -73, "80:37:73:ba:f7:d8": -48, "20:aa:4b:b8:31:c8": -81, "00:1a:1e:46:cd:10": -79, "2c:b0:5d:36:e3:b8": -81, "e0:46:9a:6d:02:ea": -84, "b4:75:0e:03:cd:66": -83, "a0:63:91:2b:9e:64": -72, "30:46:9a:a0:28:c4": -76, "e8:ed:05:55:21:10": -83, "4c:60:de:fe:e5:24": -85, "b4:75:0e:03:cd:69": -90, "d4:05:98:57:b3:10": -81, "ec:1a:59:4a:9c:ed": -83, "70:73:cb:bd:9f:b5": -76, "00:ac:e0:b8:ea:a0": -81, "a0:63:91:2b:9e:65": -55}}}

{"method": "import_measurements", "location_id": "01e25343-80a1-c736-dc6e-16914ccf92dc", "device_id": "a863c68f-9010-8afe-a12d-738ae312b7ad", "apikey": "c2d7589a472941efb29577833dd6a5fb", "data": {"c113b56e-f03c-dfc5-6ead-c458373b3f5d": {"80:37:73:ba:f7:dc": -46, "00:1a:1e:46:cd:11": -76, "f8:35:dd:0a:da:be": -84, "34:68:95:f8:25:fd": -77, "b8:3e:59:78:35:99": -87, "00:23:69:d4:47:9f": -72, "80:37:73:ba:f7:d8": -50, "20:aa:4b:b8:31:c8": -83, "e8:ed:05:55:21:15": -81, "00:1a:1e:46:cd:10": -78, "2c:b0:5d:36:e3:b8": -78, "c0:c1:c0:c3:f4:9b": -89, "20:e5:2a:9d:bc:ee": -91, "b4:75:0e:03:cd:66": -83, "a0:21:b7:b0:7c:a2": -87, "a0:63:91:2b:9e:64": -72, "30:46:9a:a0:28:c4": -73, "4c:60:de:fe:e5:24": -82, "b4:75:0e:03:cd:69": -90, "d4:05:98:57:b3:10": -81, "70:73:cb:bd:9f:b5": -76, "00:ac:e0:b8:ea:a0": -81, "a0:63:91:2b:9e:65": -60}}}

{"method": "import_measurements", "location_id": "01e25343-80a1-c736-dc6e-16914ccf92dc", "device_id": "a863c68f-9010-8afe-a12d-738ae312b7ad", "apikey": "c2d7589a472941efb29577833dd6a5fb", "data": {"c113b56e-f03c-dfc5-6ead-c458373b3f5d": {"80:37:73:ba:f7:dc": -46, "00:1a:1e:46:cd:11": -76, "f8:35:dd:0a:da:be": -84, "34:68:95:f8:25:fd": -77, "b8:3e:59:78:35:99": -87, "00:23:69:d4:47:9f": -72, "80:37:73:ba:f7:d8": -50, "20:aa:4b:b8:31:c8": -83, "e8:ed:05:55:21:15": -81, "00:1a:1e:46:cd:10": -78, "2c:b0:5d:36:e3:b8": -78, "c0:c1:c0:c3:f4:9b": -89, "20:e5:2a:9d:bc:ee": -91, "b4:75:0e:03:cd:66": -83, "a0:21:b7:b0:7c:a2": -87, "a0:63:91:2b:9e:64": -72, "30:46:9a:a0:28:c4": -73, "4c:60:de:fe:e5:24": -82, "b4:75:0e:03:cd:69": -90, "d4:05:98:57:b3:10": -81, "70:73:cb:bd:9f:b5": -76, "00:ac:e0:b8:ea:a0": -81, "a0:63:91:2b:9e:65": -60}}}


*/

var (
	TEST_USER     *database.User
	TEST_USERNAME = ""
	TEST_PASSWORD = "test1234"
	TEST_DEVICE   *database.Device
)

func checkError(t *testing.T, resp *api.Response, err error) {
	if nil != err {
		t.Error(err.Error())
	}
	if "ok" != resp.Status {
		t.Error(resp.Error)
	}
}

// Users
func TestUserCreation(t *testing.T) {
	resp, err := findapi.DoJSON(fmt.Sprintf(`{"method": "create_user", "username": "test_user_%v", "password": "test1234"}`, time.Now().UnixNano()))
	checkError(t, resp, err)
}

func TestGetUser(t *testing.T) {
	resp, err := findapi.DoJSON(fmt.Sprintf(`{"method": "get_user", "apikey": "%v"}`, TEST_USER.Apikey))
	checkError(t, resp, err)
}

func BenchmarkUserCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		findapi.DoJSON(fmt.Sprintf(`{"method": "create_user", "username": "test_user_%v", "password": "test1234"}`, time.Now().UnixNano()))
	}
}

func BenchmarkGetUSer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		findapi.DoJSON(fmt.Sprintf(`{"method": "get_user", "apikey": "%v"}`, TEST_USER.Apikey))
	}
}

// Devices
func TestDeviceCreation(t *testing.T) {
	resp, err := findapi.DoJSON(fmt.Sprintf(`{"method": "create_device", "apikey": "%v", "name": "test_device", "type": "test"}`, TEST_USER.Apikey))
	checkError(t, resp, err)
}

func TestGetDevice(t *testing.T) {
	resp, err := findapi.DoJSON(fmt.Sprintf(`{"method": "get_devices", "apikey": "%v"}`, TEST_USER.Apikey))
	checkError(t, resp, err)
}

// get_device

func BenchmarkDeviceCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		findapi.DoJSON(fmt.Sprintf(`{"method": "create_device", "apikey": "%v", "name": "test_device_%v", "type": "test"}`, TEST_USER.Apikey, time.Now().UnixNano()))
	}
}

func BenchmarkGetDevices(b *testing.B) {
	for i := 0; i < b.N; i++ {
		findapi.DoJSON(fmt.Sprintf(`{"method": "get_devices", "apikey": "%v"}`, TEST_USER.Apikey))
	}
}

// get_device

// https://golang.org/pkg/testing/#pkg-overview
func TestMain(m *testing.M) {
	// main.init parses command line args
	// set up api and database
	dbConnectionString := fmt.Sprintf("%v://%v:%v@%v:%v/%v?sslmode=disable", DATABASE_ENGINE, DATABASE_USERNAME, DATABASE_PASSWORD, DATABASE_HOST, DATABASE_PORT, DATABASE_DATABASE)
	redisAddr := fmt.Sprintf("%v:%v", REDIS_HOST, REDIS_PORT)
	findapi = api.New(dbConnectionString, "localhost:8002", redisAddr)

	// setup objects for testing
	TEST_USERNAME = fmt.Sprintf("test_user_%v", time.Now().UnixNano())
	resp, err := findapi.DoJSON(fmt.Sprintf(`{"method": "create_user", "username": "%v", "password": "%v"}`, TEST_USERNAME, TEST_PASSWORD))
	if nil != err {
		panic(err)
	}

	resp, err = findapi.DoJSON(fmt.Sprintf(`{"method": "get_user", "username": "%v"}`, TEST_USERNAME))
	if nil != err {
		panic(err)
	}
	TEST_USER = resp.Data.User

	device_name := fmt.Sprintf(`test_device_%v`, time.Now().UnixNano())
	resp, err = findapi.DoJSON(fmt.Sprintf(`{"method": "create_device", "apikey": "%v", "name": "%v", "type": "test"}`, TEST_USER.Apikey, device_name))
	if nil != err {
		panic(err)
	}

	resp, err = findapi.DoJSON(fmt.Sprintf(`{"method": "get_devices", "apikey": "%v"}`, TEST_USER.Apikey))
	if nil != err {
		panic(err)
	}
	TEST_DEVICE = resp.Data.Devices[0]
	// TODO: GET DEVICE

	// run tests
	os.Exit(m.Run())
}
