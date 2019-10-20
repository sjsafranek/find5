package main

import (
	"bytes"
	"encoding/json"
	"flag"
	// "fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/schollz/wifiscan"
	"github.com/sjsafranek/find5/findapi/lib/api"
)

const (
	DEFAULT_API_SERVER            = "http://localhost:8080/api/v1/find"
	DEFAULT_WIFI_INTERFACE string = "wlan0"
	DEFAULT_SENSOR_ID      string = ""
	DEFAULT_DEVICE_ID      string = ""
	DEFAULT_LOCATION_ID    string = ""
	DEFAULT_APIKEY         string = ""
)

var (
	WIFI_INTERFACE string = DEFAULT_WIFI_INTERFACE
	SENSOR_ID      string = DEFAULT_SENSOR_ID
	DEVICE_ID      string = DEFAULT_DEVICE_ID
	LOCATION_ID    string = DEFAULT_LOCATION_ID
	APIKEY         string = DEFAULT_APIKEY
	API_SERVER     string = DEFAULT_API_SERVER
	client                = http.Client{
		Timeout: time.Duration(15 * time.Second),
	}
)

func send(apiRequest api.Request) (string, error) {
	requestBody, err := json.Marshal(apiRequest)
	if nil != err {
		return "", err
	}

	log.Println("[OUT]", string(requestBody))

	request, err := http.NewRequest("POST", API_SERVER, bytes.NewBuffer(requestBody))
	if nil != err {
		return "", err
	}

	request.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(request)
	if nil != err {
		return "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if nil != err {
		return "", err
	}

	log.Println("[IN] ", body)

	return string(body), nil
}

func scan(wifiInterface string) (map[string]float64, error) {
	data := make(map[string]float64)
	wifis, err := wifiscan.Scan(wifiInterface)
	if err != nil {
		return data, err
	}
	for _, w := range wifis {
		data[w.SSID] = float64(w.RSSI)
	}
	return data, nil
}

func init() {

	flag.StringVar(&API_SERVER, "api_server", DEFAULT_API_SERVER, "Api server")
	flag.StringVar(&WIFI_INTERFACE, "wifi_interface", DEFAULT_WIFI_INTERFACE, "")
	flag.StringVar(&SENSOR_ID, "sensor_id", DEFAULT_SENSOR_ID, "")
	flag.StringVar(&DEVICE_ID, "device_id", DEFAULT_DEVICE_ID, "")
	flag.StringVar(&LOCATION_ID, "location_id", DEFAULT_LOCATION_ID, "")
	flag.StringVar(&APIKEY, "apikey", DEFAULT_APIKEY, "")
	flag.Parse()
}

func main() {

	for {
		data, err := scan(WIFI_INTERFACE)
		if nil != err {
			panic(err)
		}

		var apiRequest api.Request
		apiRequest.Method = "import_measurements"
		apiRequest.Apikey = APIKEY
		apiRequest.DeviceId = DEVICE_ID
		apiRequest.LocationId = LOCATION_ID
		apiRequest.Data = make(map[string]map[string]float64)
		apiRequest.Data[SENSOR_ID] = data
		apiRequest.Timestamp = time.Now()

		result, err := send(apiRequest)
		if nil != err {
			log.Println(err)
			log.Println("TODO: store request for future use....")
			time.Sleep(5 * time.Second)
			continue
		}
		log.Println(result)

		time.Sleep(5 * time.Second)
	}

}
