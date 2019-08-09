package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/schollz/find4/server/main/src/api"
	"github.com/schollz/find4/server/main/src/database"
	"github.com/schollz/find4/server/main/src/models"
	// "github.com/schollz/find4/server/main/src/mqtt"
	"github.com/schollz/utils"
)

// Port defines the public port
var Port = "8003"
var UseSSL = false
var UseMQTT = false
var MinimumPassive = -1

// Run will start the server listening on the specified port
func Run() (err error) {
	defer logger.Flush()

	// if UseMQTT {
	// 	// setup MQTT
	// 	err = mqtt.Setup()
	// 	if err != nil {
	// 		logger.Warn(err)
	// 	}
	// 	logger.Debug("setup mqtt")
	// }

	logger.Debug("current families: ", database.GetFamilies())

	// setup gin server
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	// Standardize logs
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")
	r.Use(middleWareHandler(), gin.Recovery(), gzip.Gzip(gzip.DefaultCompression))
	// r.Use(middleWareHandler(), gin.Recovery())
	r.HEAD("/", func(c *gin.Context) { // handler for the uptime robot
		c.String(http.StatusOK, "OK")
	})
	r.GET("/", func(c *gin.Context) { // handler for the uptime robot
		c.HTML(http.StatusOK, "login.tmpl", gin.H{
			"Message": "",
		})
	})
	r.POST("/", func(c *gin.Context) {
		family := c.PostForm("inputFamily")
		// db, err := GetDatabase(c.Param("family"))
		// if err == nil {
		// c.Redirect(http.StatusMovedPermanently, "/view/dashboard/"+family)
		// } else {
		c.HTML(http.StatusOK, "login.tmpl", gin.H{
			"Message": template.HTML(fmt.Sprintf(`Family '%s' does not exist. Follow <a href="https://www.internalpositioning.com/doc/tracking_your_phone.md" target="_blank">these instructions</a> to get started.`, family)),
		})
		// }

	})
	r.DELETE("/api/v1/database/:family", func(c *gin.Context) {
		err := DeleteDatabase(c.Param("family"))
		if err == nil {
			c.JSON(200, gin.H{"success": true, "message": "deleted " + c.Param("family")})
		} else {
			c.JSON(200, gin.H{"success": false, "message": err.Error()})
		}

	})
	r.DELETE("/api/v1/location/:family/:location", func(c *gin.Context) {
		db, err := GetDatabase(c.Param("family"))
		if err == nil {
			err = db.DeleteLocation(c.Param("location"))
			if err == nil {
				c.JSON(200, gin.H{"success": true, "message": "deleted location '" + c.Param("location") + "' for " + c.Param("family")})
				return
			}
		}
		c.JSON(200, gin.H{"success": false, "message": err.Error()})
	})
	r.GET("/view/location/:family/:device", func(c *gin.Context) {
		family := c.Param("family")
		device := c.Param("device")
		c.HTML(http.StatusOK, "location.tmpl", gin.H{
			"Family":   family,
			"Device":   device,
			"FamilyJS": template.JS(family),
			"DeviceJS": template.JS(device),
		})
	})
	r.GET("/api/v1/database/:family", func(c *gin.Context) {
		db, err := GetDatabase(c.Param("family"))
		if err == nil {
			var dumped string
			dumped, err = db.Dump()
			if err == nil {
				c.String(200, dumped)
				return
			}
		}
		c.JSON(200, gin.H{"success": false, "message": err.Error()})
	})
	r.GET("/view/dashboard/:family", func(c *gin.Context) {
		type LocEff struct {
			Name           string
			Total          int64
			PercentCorrect int64
		}
		type Efficacy struct {
			AccuracyBreakdown   []LocEff
			LastCalibrationTime time.Time
			TotalCount          int64
			PercentCorrect      int64
		}
		type DeviceTable struct {
			ID           string
			Name         string
			LastLocation string
			LastSeen     time.Time
			Probability  int64
			ActiveTime   int64
		}

		family := c.Param("family")
		err := func(family string) (err error) {
			startTime := time.Now()
			var errorMessage string

			db, err := GetDatabase(family)
			if err != nil {
				err = errors.Wrap(err, "You need to add learning data first")
				return
			}
			var efficacy Efficacy

			minutesAgoInt := 60
			millisecondsAgo := int64(minutesAgoInt * 60 * 1000)
			sensors, err := db.GetSensorFromGreaterTime(millisecondsAgo)
			logger.Debugf("[%s] got sensor from greater time %s", family, time.Since(startTime))
			devicesToCheckMap := make(map[string]struct{})
			for _, sensor := range sensors {
				devicesToCheckMap[sensor.Device] = struct{}{}
			}
			// get list of devices I care about
			devicesToCheck := make([]string, len(devicesToCheckMap))
			i := 0
			for device := range devicesToCheckMap {
				devicesToCheck[i] = device
				i++
			}
			logger.Debugf("[%s] found %d devices to check", family, len(devicesToCheck))

			logger.Debugf("[%s] getting device counts", family)
			deviceCounts, err := db.GetDeviceCountsFromDevices(devicesToCheck)
			if err != nil {
				err = errors.Wrap(err, "could not get devices")
				return
			}

			deviceList := make([]string, len(deviceCounts))
			i = 0
			for device := range deviceCounts {
				if deviceCounts[device] > 2 {
					deviceList[i] = device
					i++
				}
			}
			deviceList = deviceList[:i]
			jsonDeviceList, _ := json.Marshal(deviceList)
			logger.Debugf("found %d devices", len(deviceList))

			logger.Debugf("[%s] getting locations", family)
			locationList, err := db.GetLocations()
			if err != nil {
				logger.Warn("could not get locations")
			}
			jsonLocationList, _ := json.Marshal(locationList)
			logger.Debugf("found %d locations", len(locationList))

			logger.Debugf("[%s] total learned count", family)
			efficacy.TotalCount, err = db.TotalLearnedCount()
			if err != nil {
				logger.Warn("could not get TotalLearnedCount")
			}

			// var percentFloat64 float64
			// err = db.Get("PercentCorrect", &percentFloat64)
			// if err != nil {
			// 	logger.Warn("No learning data available")
			// }
			// efficacy.PercentCorrect = int64(100 * percentFloat64)
			// err = db.Get("LastCalibrationTime", &efficacy.LastCalibrationTime)
			// if err != nil {
			// 	logger.Warn("could not get LastCalibrationTime")
			// }
			// var accuracyBreakdown map[string]float64
			// err = db.Get("AccuracyBreakdown", &accuracyBreakdown)
			// if err != nil {
			// 	logger.Warn("could not get AccuracyBreakdown")
			// }
			// var confusionMetrics map[string]map[string]models.BinaryStats
			// err = db.Get("AlgorithmEfficacy", &confusionMetrics)
			// if err != nil {
			// 	logger.Warn("could not get AlgorithmEfficacy")
			// }

			// DEBUGING
			calibration, err := db.GetCalibration()
			if err != nil {
				logger.Warn("could not get calibration")
			}
			percentFloat64 := calibration.PercentCorrect
			efficacy.PercentCorrect = int64(100 * calibration.PercentCorrect)
			efficacy.LastCalibrationTime = calibration.CalibrationTime
			accuracyBreakdown := calibration.AccuracyBreakdown
			// confusionMetrics := calibration.AlgorithmEfficacy
			//.end

			logger.Debugf("[%s] getting location count", family)
			locationCounts, err := db.GetLocationCounts()
			if err != nil {
				logger.Warn("could not get location counts")
			}
			logger.Debugf("[%s] locations: %+v", family, locationCounts)

			efficacy.AccuracyBreakdown = make([]LocEff, len(accuracyBreakdown))
			i = 0
			for key := range accuracyBreakdown {
				l := LocEff{Name: strings.Title(key)}
				l.PercentCorrect = int64(100 * accuracyBreakdown[key])
				l.Total = int64(locationCounts[key])
				efficacy.AccuracyBreakdown[i] = l
				i++
			}
			var rollingData models.ReverseRollingData
			errRolling := db.Get("ReverseRollingData", &rollingData)
			passiveTable := []DeviceTable{}
			scannerList := []string{}
			if errRolling == nil {
				passiveTable = make([]DeviceTable, len(rollingData.DeviceLocation))
				i := 0
				for device := range rollingData.DeviceLocation {
					s, errOpen := db.GetLatest(device)
					if errOpen != nil {
						continue
					}
					passiveTable[i].Name = device
					passiveTable[i].LastLocation = rollingData.DeviceLocation[device]
					passiveTable[i].LastSeen = time.Unix(0, s.Timestamp*1000000).UTC()
					i++
				}
				sensors, errGet := db.GetSensorFromGreaterTime(60000 * 15)
				if errGet == nil {
					allScanners := make(map[string]struct{})
					for _, s := range sensors {
						for sensorType := range s.Sensors {
							for scanner := range s.Sensors[sensorType] {
								allScanners[scanner] = struct{}{}
							}
						}
					}
					scannerList = make([]string, len(allScanners))
					i = 0
					for scanner := range allScanners {
						scannerList[i] = scanner
						i++
					}
				}
			}

			logger.Debugf("[%s] getting by_locations", family)
			byLocations, err := api.GetByLocation(db, family, 15, false, 3, 0, 0, deviceCounts)
			if err != nil {
				logger.Warn(err)
			}

			logger.Debugf("[%s] creating device table", family)
			table := []DeviceTable{}
			for _, byLocation := range byLocations {
				for _, device := range byLocation.Devices {
					table = append(table, DeviceTable{
						ID:           utils.Hash(device.Device),
						Name:         device.Device,
						LastLocation: byLocation.Location,
						LastSeen:     device.Timestamp,
						Probability:  int64(device.Probability * 100),
						ActiveTime:   int64(device.ActiveMins),
					})
				}
			}

			if err != nil {
				errorMessage = err.Error()
			} else if percentFloat64 == 0 {
				errorMessage = "No learning data available, see the documentation for how to get started with learning. "
			}

			// WHY ISNT THIS AUTOMATICALLY DONE?
			if efficacy.LastCalibrationTime.IsZero() {
				errorMessage += "You need to calibrate, press the calibration button."
			}

			c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
				"Family":         family,
				"FamilyJS":       template.JS(family),
				"Efficacy":       efficacy,
				"Devices":        table,
				"ErrorMessage":   errorMessage,
				"PassiveDevices": passiveTable,
				"DeviceList":     template.JS(jsonDeviceList),
				"LocationList":   template.JS(jsonLocationList),
				"Scanners":       scannerList,
				"PercentCorrect": percentFloat64,
				"UseMQTT":        UseMQTT,
				"MQTTServer":     os.Getenv("MQTT_EXTERNAL"),
				"MQTTPort":       os.Getenv("MQTT_PORT"),
			})
			err = nil
			logger.Debugf("[%s] rendered dashboard in %s", family, time.Since(startTime))
			return
		}(family)
		if err != nil {
			logger.Warn(err)
			c.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
				"Family":       family,
				"FamilyJS":     template.JS(family),
				"ErrorMessage": err.Error(),
				"Efficacy":     Efficacy{},
			})
		}
	})
	r.OPTIONS("/api/v1/devices/*family", func(c *gin.Context) { c.String(200, "OK") })
	r.GET("/api/v1/devices/*family", handlerApiV1Devices)
	r.OPTIONS("/api/v1/location/:family/*device", func(c *gin.Context) { c.String(200, "OK") })
	r.GET("/api/v1/location/:family/*device", handlerApiV1Location)
	r.OPTIONS("/api/v1/locations/:family", func(c *gin.Context) { c.String(200, "OK") })
	r.GET("/api/v1/locations/:family", handlerApiV1Locations)
	r.OPTIONS("/api/v1/location_basic/:family/*device", func(c *gin.Context) { c.String(200, "OK") })
	r.GET("/api/v1/location_basic/:family/*device", handlerApiV1LocationSimple)
	r.OPTIONS("/api/v1/by_location/:family", func(c *gin.Context) { c.String(200, "OK") })
	r.GET("/api/v1/by_location/:family", handlerApiV1ByLocation)
	r.OPTIONS("/api/v1/calibrate/*family", func(c *gin.Context) { c.String(200, "OK") })
	r.GET("/api/v1/calibrate/*family", handlerApiV1Calibrate)
	r.OPTIONS("/api/v1/settings/passive", func(c *gin.Context) { c.String(200, "OK") })
	r.POST("/api/v1/settings/passive", handlerReverseSettings)
	r.OPTIONS("/api/v1/efficacy/:family", func(c *gin.Context) { c.String(200, "OK") })
	r.GET("/api/v1/efficacy/:family", handlerEfficacy)
	r.GET("/ping", ping)
	r.GET("/now", handlerNow)
	r.GET("/test", handleTest)
	r.GET("/ws", wshandler) // handler for the web sockets (see websockets.go)
	// if UseMQTT {
	// 	r.GET("/api/v1/mqtt/:family", handlerMQTT) // handler for setting MQTT
	// }
	r.POST("/data", handlerData)             // typical data handler
	r.POST("/classify", handlerDataClassify) // classify a fingerprint
	r.POST("/passive", handlerReverse)       // typical data handler
	r.POST("/learn", handlerFIND)            // backwards-compatible with FIND for learning
	r.POST("/track", handlerFIND)            // backwards-compatible with FIND for tracking
	logger.Infof("Running on 0.0.0.0:%s", Port)

	err = r.Run(":" + Port) // listen and serve on 0.0.0.0:8080
	return
}

func replace(input, from, to string) string {
	return strings.Replace(input, from, to, -1)
}

func ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func handleTest(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func handlerApiV1Devices(c *gin.Context) {
	err := func(c *gin.Context) (err error) {
		family := strings.TrimSpace(c.Param("family")[1:])
		db, err := GetDatabase(family)
		if err != nil {
			return
		}
		s, err := db.GetDevices()
		if err != nil {
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "got devices", "success": true, "devices": s})
		return
	}(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": false})
	}
}

func handlerApiV1Locations(c *gin.Context) {
	type Location struct {
		Device     string                    `json:"device"`
		Sensors    models.SensorData         `json:"sensors"`
		Prediction models.LocationPrediction `json:"prediction"`
	}

	locations, err := func(c *gin.Context) (locations []Location, err error) {
		family := strings.TrimSpace(c.Param("family"))

		db, err := GetDatabase(family)
		if err != nil {
			return
		}
		devices, err := db.GetDevices()
		if err != nil {
			return
		}
		locations = make([]Location, len(devices))
		logger.Debugf("[%s] getting information for %d devices", family, len(devices))
		for i, device := range devices {
			logger.Debugf("[%s] getting prediction for %s", family, device)
			locations[i] = Location{Device: device}
			locations[i].Sensors, err = db.GetLatest(device)
			if err != nil {
				continue
			}
			predictions, err := db.GetPrediction(locations[i].Sensors.Timestamp)
			if err == nil && len(predictions) > 0 {
				locations[i].Prediction = predictions[0]
			} else {
				analysis, err := api.AnalyzeSensorData(db, locations[i].Sensors)
				if err != nil {
					continue
				}
				if len(analysis.Guesses) > 0 {
					locations[i].Prediction = analysis.Guesses[0]
				}
			}
		}

		return
	}(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": err == nil})
	} else {

		c.JSON(http.StatusOK, gin.H{"message": "got locations", "success": err == nil, "locations": locations})
	}
}

func handlerEfficacy(c *gin.Context) {
	type Efficacy struct {
		AccuracyBreakdown   map[string]float64                       `json:"accuracy_breakdown"`
		ConfusionMetrics    map[string]map[string]models.BinaryStats `json:"confusion_metrics"`
		LastCalibrationTime time.Time                                `json:"last_calibration_time"`
	}
	efficacy, err := func(c *gin.Context) (efficacy Efficacy, err error) {
		family := strings.TrimSpace(c.Param("family"))

		db, err := GetDatabase(family)
		if err != nil {
			return
		}

		// err = db.Get("LastCalibrationTime", &efficacy.LastCalibrationTime)
		// if err != nil {
		// 	err = errors.Wrap(err, "could not get LastCalibrationTime")
		// 	return
		// }
		// err = db.Get("AccuracyBreakdown", &efficacy.AccuracyBreakdown)
		// if err != nil {
		// 	err = errors.Wrap(err, "could not get AccuracyBreakdown")
		// 	return
		// }
		// err = db.Get("AlgorithmEfficacy", &efficacy.ConfusionMetrics)
		// if err != nil {
		// 	err = errors.Wrap(err, "could not get AlgorithmEfficacy")
		// 	return
		// }

		// DEBUGING
		calibration, err := db.GetCalibration()
		if err != nil {
			logger.Warn("could not get calibration")
			err = errors.Wrap(err, "could not get calibration")
			return
		}
		// percentFloat64 := calibration.PercentCorrect
		// efficacy.PercentCorrect = int64(100 * calibration.PercentCorrect)
		efficacy.LastCalibrationTime = calibration.CalibrationTime
		efficacy.AccuracyBreakdown = calibration.AccuracyBreakdown
		efficacy.ConfusionMetrics = calibration.AlgorithmEfficacy
		//.end

		return
	}(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": err == nil})
	} else {

		c.JSON(http.StatusOK, gin.H{"message": "got stats", "success": err == nil, "efficacy": efficacy})
	}
}

func handlerApiV1ByLocation(c *gin.Context) {
	locations, err := func(c *gin.Context) (byLocations []models.ByLocation, err error) {
		family := strings.TrimSpace(c.Param("family"))
		minutesAgo := strings.TrimSpace(c.DefaultQuery("history", "120"))
		showRandomized := c.DefaultQuery("randomized", "1") == "1"
		activeMinsThreshold, err := strconv.Atoi(c.DefaultQuery("active_mins", "0"))
		if err != nil {
			return
		}
		minScanners, err := strconv.Atoi(c.DefaultQuery("num_scanners", "0"))
		if err != nil {
			return
		}
		minProbability, err := strconv.ParseFloat(c.DefaultQuery("probability", "0"), 64)
		if err != nil {
			return
		}
		minutesAgoInt, err := strconv.Atoi(minutesAgo)
		if err != nil {
			return
		}

		db, err := GetDatabase(family)
		if err != nil {
			return
		}

		byLocations, err = api.GetByLocation(db, family, minutesAgoInt, showRandomized, activeMinsThreshold, minScanners, minProbability, make(map[string]int))
		return
	}(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": err == nil})
	} else {

		c.JSON(http.StatusOK, gin.H{"message": "got locations", "success": err == nil, "locations": locations})
	}
}

func handlerApiV1Location(c *gin.Context) {
	s, analysis, err := func(c *gin.Context) (s models.SensorData, analysis models.LocationAnalysis, err error) {
		family := strings.TrimSpace(c.Param("family"))
		device := strings.TrimSpace(c.Param("device")[1:])

		db, err := GetDatabase(family)
		if err != nil {
			return
		}
		s, err = db.GetLatest(device)
		if err != nil {
			return
		}
		analysis, err = api.AnalyzeSensorData(db, s)
		if err != nil {
			err = api.Calibrate(db, family, true)
			if err != nil {
				logger.Warn(err)
				return
			}
		}
		return
	}(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": err == nil})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "got location", "success": err == nil, "sensors": s, "analysis": analysis})
	}
}

func handlerApiV1LocationSimple(c *gin.Context) {
	s, analysis, err := func(c *gin.Context) (s models.SensorData, analysis models.LocationAnalysis, err error) {
		family := strings.TrimSpace(c.Param("family"))
		device := strings.TrimSpace(c.Param("device")[1:])
		logger.Debugf("[%s] getting location for %s", family, device)

		db, err := GetDatabase(family)
		if err != nil {
			return
		}
		s, err = db.GetLatest(device)
		if err != nil {
			return
		}
		analysis, err = api.AnalyzeSensorData(db, s)
		if err != nil {
			err = api.Calibrate(db, family, true)
			if err != nil {
				logger.Warn(err)
				return
			}
		}
		return
	}(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": err == nil})
	} else {
		simpleLocation := struct {
			Location        string  `json:"loc"`
			Probability     float64 `json:"prob"`
			LastSeenTimeAgo int64   `json:"seen"`
		}{
			Location:        analysis.Guesses[0].Location,
			Probability:     analysis.Guesses[0].Probability,
			LastSeenTimeAgo: time.Now().UTC().UnixNano()/int64(time.Second) - (s.Timestamp / 1000),
		}
		c.JSON(http.StatusOK, gin.H{"message": "ok", "success": err == nil, "data": simpleLocation})
	}
}

func handlerApiV1Calibrate(c *gin.Context) {
	family := strings.TrimSpace(c.Param("family")[1:])
	var err error
	if family == "" {
		err = errors.New("invalid family")
	} else {
		db, err := GetDatabase(family)
		if err != nil {
			return
		}
		err = api.Calibrate(db, family, true)
	}
	message := "calibrated data"
	if err != nil {
		message = err.Error()
	}
	c.JSON(http.StatusOK, gin.H{"message": message, "success": err == nil})
}

/*
func handlerMQTT(c *gin.Context) {
	message, err := func(c *gin.Context) (message string, err error) {
		family := strings.TrimSpace(c.Param("family"))
		if family == "" {
			err = errors.New("invalid family")
			return
		}
		passphrase, err := mqtt.AddFamily(family)
		if err != nil {
			return
		}
		message = fmt.Sprintf("Added '%s' for mqtt. Your passphrase is '%s'", family, passphrase)
		return
	}(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": err == nil})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": message, "success": err == nil})
	}
	return
}
*/

func sendOutLocation(family, device string) (s models.SensorData, analysis models.LocationAnalysis, err error) {
	db, err := GetDatabase(family)
	if err != nil {
		return
	}
	s, err = db.GetLatest(device)
	if err != nil {
		return
	}
	analysis, err = sendOutData(s)
	if err != nil {
		return
	}
	analysis, err = api.AnalyzeSensorData(db, s)
	if err != nil {
		err = api.Calibrate(db, family, true)
		if err != nil {
			logger.Warn(err)
			return
		}
	}
	return
}

func handlerNow(c *gin.Context) {
	c.String(200, strconv.Itoa(int(time.Now().UTC().UnixNano()/int64(time.Millisecond))))
}

func handlerData(c *gin.Context) {
	message, err := func(c *gin.Context) (message string, err error) {
		justSave := c.DefaultQuery("justsave", "0") == "1"
		var d models.SensorData
		err = c.BindJSON(&d)
		if err != nil {
			message = d.Family
			err = errors.Wrap(err, "problem binding data")
			return
		}

		err = d.Validate()
		if err != nil {
			message = d.Family
			err = errors.Wrap(err, "problem validating data")
			return
		}

		// process data
		err = processSensorData(d, justSave)
		if err != nil {
			message = d.Family
			return
		}
		message = "inserted data"

		logger.Debugf("[%s] /data", d.Family)
		return
	}(c)

	if err != nil {
		logger.Debugf("[%s] problem parsing: %s", message, err.Error())
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": false})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": message, "success": true})
	}
}

func handlerDataClassify(c *gin.Context) {
	aidata, message, err := func(c *gin.Context) (aidata models.LocationAnalysis, message string, err error) {
		var d models.SensorData
		err = c.BindJSON(&d)
		if err != nil {
			err = errors.Wrap(err, "problem binding data")
			return
		}

		err = d.Validate()
		if err != nil {
			err = errors.Wrap(err, "problem validating data")
			return
		}

		// process data
		err = processSensorData(d, true)
		if err != nil {
			return
		}

		db, err := GetDatabase(d.Family)
		if err != nil {
			return
		}

		aidata, err = api.AnalyzeSensorData(db, d)
		logger.Debugf("[%s] /data %+v", d.Family, d)
		message = "classified data"
		return
	}(c)

	if err != nil {
		logger.Debugf("problem parsing: %s", err.Error())
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": false, "analysis": nil})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": message, "success": true, "analysis": aidata})
	}
}

func handlerReverseSettings(c *gin.Context) {
	message, err := func(c *gin.Context) (message string, err error) {
		// bind sensor data
		type ReverseSettings struct {
			// Minimum number of passive
			MinimumPassive int `json:"minimum_passive"`
			// Timespan of window
			Window int64 `json:"window"`
			// Family is a group of devices
			Family string `json:"family" binding:"required"`
			// Device are unique within a family
			Device string `json:"device"`
			// Location is optional, used for designating learning
			Location string `json:"location"`
			// Latitude
			Latitude float64 `json:"lat"`
			// Longitude
			Longitude float64 `json:"lon"`
			// Altitude
			Altitude float64 `json:"alt"`
		}
		var d ReverseSettings
		err = c.BindJSON(&d)
		if err != nil {
			err = errors.Wrap(err, "could not bind json")
			return
		}
		d.Family = strings.TrimSpace(strings.ToLower(d.Family))
		d.Device = strings.TrimSpace(strings.ToLower(d.Device))
		d.Location = strings.TrimSpace(strings.ToLower(d.Location))

		// open database
		db, err := GetDatabase(c.Param("family"))
		if err != nil {
			return
		}

		var rollingData models.ReverseRollingData
		err = db.Get("ReverseRollingData", &rollingData)
		if err != nil {
			rollingData = models.ReverseRollingData{
				Family:         d.Family,
				DeviceLocation: make(map[string]string),
				DeviceGPS:      make(map[string]models.GPS),
				TimeBlock:      90 * time.Second,
			}
		}
		if rollingData.TimeBlock.Seconds() == 0 {
			rollingData.TimeBlock = 90 * time.Second
		}

		// set tracking information
		if d.Device != "" {
			if d.Location != "" {
				message = fmt.Sprintf("Set location to '%s' for %s for learning with device '%s'", d.Location, d.Family, d.Device)
				rollingData.DeviceLocation[d.Device] = d.Location
				if d.Latitude != 0 && d.Longitude != 0 {
					rollingData.DeviceGPS[d.Device] = models.GPS{
						Latitude:  d.Latitude,
						Longitude: d.Longitude,
						Altitude:  d.Altitude,
					}
				}
			} else {
				message = fmt.Sprintf("switched to tracking for %s", d.Family)
				delete(rollingData.DeviceLocation, d.Device)
			}
			message += ". "
		}
		message += fmt.Sprintf("Now learning on %d devices: %+v", len(rollingData.DeviceLocation), rollingData.DeviceLocation)

		// set time block information
		if d.Window > 0 {
			rollingData.TimeBlock = time.Duration(d.Window) * time.Second
		}
		message += fmt.Sprintf("with time block of %2.0f seconds", rollingData.TimeBlock.Seconds())

		if d.MinimumPassive != 0 {
			rollingData.MinimumPassive = d.MinimumPassive
			message += fmt.Sprintf(" and set minimum passive to %d", rollingData.MinimumPassive)
		}

		err = db.Set("ReverseRollingData", rollingData)
		logger.Debugf("[%s] %s", d.Family, message)
		return
	}(c)

	if err != nil {
		logger.Warn(err)
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": false})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": message, "success": true})
	}
}

func handlerReverse(c *gin.Context) {
	message, err := func(c *gin.Context) (message string, err error) {
		// bind sensor data
		var d models.SensorData
		err = c.BindJSON(&d)
		if err != nil {
			logger.Warn(err)
			return
		}

		// validate sensor data
		err = d.Validate()
		if err != nil {
			logger.Warn(err)
			return
		}

		if d.Location != "" {
			logger.Debugf("[%s] entered passive fingerprint for %s at %s", d.Family, d.Device, d.Location)
		} else {
			logger.Debugf("[%s] entered passive fingerprint for %s", d.Family, d.Device)
		}

		// open database
		db, err := GetDatabase(c.Param("family"))
		if err != nil {
			return
		}

		var rollingData models.ReverseRollingData
		err = db.Get("ReverseRollingData", &rollingData)
		if err != nil {
			// defaults
			rollingData = models.ReverseRollingData{
				Family:         d.Family,
				DeviceLocation: make(map[string]string),
				TimeBlock:      90 * time.Second,
			}
		}
		if rollingData.TimeBlock.Seconds() == 0 {
			rollingData.TimeBlock = 90 * time.Second
		}

		if !rollingData.HasData {
			rollingData.Timestamp = time.Now().UTC()
			rollingData.Datas = []models.SensorData{}
			rollingData.HasData = true
		}
		if len(d.Sensors) == 0 {
			err = errors.New("no fingerprints")
			return
		}

		rollingData.Datas = append(rollingData.Datas, d)
		numFingerprints := 0
		for sensor := range d.Sensors {
			numFingerprints += len(d.Sensors[sensor])
		}
		err = db.Set("ReverseRollingData", rollingData)
		message = fmt.Sprintf("inserted %d fingerprints for %s", numFingerprints, d.Family)

		if err == nil {
			go parseRollingData(d.Family)
		}
		return
	}(c)

	if err != nil {
		logger.Warn(err)
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": false})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": message, "success": true})
	}

}

func parseRollingData(family string) (err error) {
	db, err := GetDatabase(family)
	if err != nil {
		return
	}

	var rollingData models.ReverseRollingData
	err = db.Get("ReverseRollingData", &rollingData)
	if err != nil {
		return
	}

	sensorMap := make(map[string]models.SensorData)
	if rollingData.HasData && time.Since(rollingData.Timestamp) > rollingData.TimeBlock {
		logger.Debugf("[%s] New data arrived %s", family, time.Since(rollingData.Timestamp))
		// merge data
		for _, data := range rollingData.Datas {
			for sensor := range data.Sensors {
				for mac := range data.Sensors[sensor] {
					rssi := data.Sensors[sensor][mac]
					trackedDeviceName := sensor + "-" + mac
					if _, ok := sensorMap[trackedDeviceName]; !ok {
						location := ""
						// if there is a device+location in map, then it is currently doing learning
						if loc, hasMac := rollingData.DeviceLocation[trackedDeviceName]; hasMac {
							location = loc
						}
						var gps models.GPS
						if g, hasMac := rollingData.DeviceGPS[trackedDeviceName]; hasMac {
							gps = g
						}
						sensorMap[trackedDeviceName] = models.SensorData{
							Family:    family,
							Device:    trackedDeviceName,
							Timestamp: time.Now().UTC().UnixNano() / int64(time.Millisecond),
							Sensors:   make(map[string]map[string]interface{}),
							Location:  location,
							GPS:       gps,
						}
						time.Sleep(10 * time.Millisecond)
						sensorMap[trackedDeviceName].Sensors[sensor] = make(map[string]interface{})
					}
					sensorMap[trackedDeviceName].Sensors[sensor][data.Device+"-"+sensor] = rssi
				}
			}
		}
		rollingData.HasData = false
	}
	db.Set("ReverseRollingData", rollingData)
	for sensor := range sensorMap {
		logger.Debugf("[%s] reverse sensor data: %+v", family, sensorMap[sensor])
		numPassivePoints := 0
		for sensorType := range sensorMap[sensor].Sensors {
			numPassivePoints += len(sensorMap[sensor].Sensors[sensorType])
		}
		if numPassivePoints < rollingData.MinimumPassive {
			logger.Debugf("[%s] skipped saving reverse sensor data for %s, not enough points (< %d)", family, sensor, rollingData.MinimumPassive)
			continue
		}
		err := processSensorData(sensorMap[sensor])
		if err != nil {
			logger.Warnf("[%s] problem saving: %s", family, err.Error())
		}
		logger.Debugf("[%s] saved reverse sensor data for %s", family, sensor)
	}

	return
}

func handlerFIND(c *gin.Context) {
	var j models.FINDFingerprint
	var err error
	var message string
	err = c.BindJSON(&j)
	if err == nil {
		if c.Request.URL.Path == "/track" {
			j.Location = ""
		}
		d := j.Convert()
		err2 := processSensorData(d)
		if err2 == nil {
			message = "inserted data"
		} else {
			err = err2
		}
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": err.Error(), "success": false})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": message, "success": true})
	}
}

func processSensorData(p models.SensorData, justSave ...bool) (err error) {
	db, err := GetDatabase(p.Family)
	if err != nil {
		return
	}

	err = api.SaveSensorData(db, p)
	if err != nil {
		return
	}

	if len(justSave) > 0 && justSave[0] {
		return
	}
	go sendOutData(p)
	return
}

func sendOutData(p models.SensorData) (analysis models.LocationAnalysis, err error) {
	db, err := GetDatabase(p.Family)
	if err != nil {
		return
	}

	analysis, _ = api.AnalyzeSensorData(db, p)
	if len(analysis.Guesses) == 0 {
		err = errors.New("no guesses")
		return
	}
	type Payload struct {
		Sensors  models.SensorData           `json:"sensors"`
		Guesses  []models.LocationPrediction `json:"guesses"`
		Location string                      `json:"location"` // FIND backwards-compatability
		Time     int64                       `json:"time"`     // FIND backwards-compatability
	}
	payload := Payload{
		Sensors:  p,
		Guesses:  analysis.Guesses,
		Location: analysis.Guesses[0].Location,
		Time:     p.Timestamp,
	}
	bTarget, err := json.Marshal(payload)
	if err != nil {
		return
	}

	// logger.Debugf("sending data over websockets (%s/%s):%s", p.Family, p.Device, bTarget)
	SendMessageOverWebsockets(p.Family, p.Device, bTarget)
	SendMessageOverWebsockets(p.Family, "all", bTarget)

	// if UseMQTT {
	// 	logger.Debugf("[%s] sending data over mqtt (%s)", p.Family, p.Device)
	// 	mqtt.Publish(p.Family, p.Device, string(bTarget))
	// }
	return
}

func middleWareHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now().UTC()
		// Add base headers
		addCORS(c)
		// Run next function
		c.Next()
		// Log request
		logger.Infof("%v %v %v %s", c.Request.RemoteAddr, c.Request.Method, c.Request.URL, time.Since(t))
	}
}

func addCORS(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Max")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
}
