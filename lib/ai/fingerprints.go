package ai

//
// import (
// 	// "time"
//
// 	"github.com/schollz/find4/server/main/src/database"
// 	"github.com/schollz/find4/server/main/src/models"
// )
//
// var (
// 	calibration_queue chan func()
// )
//
// // func init() {
// // 	// make queue length of 2 to block channel
// // 	// this will result in rate limiting of AI calibrations.
// // 	calibration_queue = make(chan func(), 10)
// // 	// Spawn goroutines to calibrate database
// // 	go calibrationWorker()
// // 	go calibrationWorker()
// // }
//
// // SaveSensorData will add sensor data to the database
// func SaveSensorData(db *database.Database, p models.SensorData) (err error) {
// 	err = p.Validate()
// 	if err != nil {
// 		return
// 	}
//
// 	err = db.AddSensor(p)
// 	if p.GPS.Longitude != 0 && p.GPS.Latitude != 0 {
// 		db.SetGPS(p)
// 	}
//
// 	if err != nil {
// 		return
// 	}
//
// 	return
// }
//
// // SavePrediction will add sensor data to the database
// func SavePrediction(db *database.Database, s models.SensorData, p models.LocationAnalysis) (err error) {
// 	err = db.AddPrediction(s.Timestamp, p.Guesses)
// 	return
// }
//
// //
// // // DatabaseWorker monitors database for changes and schedules AI calibration.
// // func DatabaseWorker(db *database.Database, family string) {
// // 	// defend against historic database inserts
// // 	var last_sensor_insert_timestamp time.Time
// // 	var last_sensor_count int
// //
// // 	// loop
// // 	for {
// // 		should_calibrate := false
// //
// // 		// compare calibration timestamp and last sensor timestamp
// // 		// defend against historic inserts
// // 		var last_calibration_time time.Time
// // 		// TODO
// // 		//  - SELECT FROM calibration TABLE
// // 		err := db.Get("LastCalibrationTime", &last_calibration_time)
// // 		if nil != err {
// // 			logger.Error(err)
// // 			should_calibrate = true
// // 		}
// //
// // 		ts, err := db.GetLastSensorInsertTimeWithLocationId()
// // 		if nil != err {
// // 			logger.Error(err)
// // 			should_calibrate = true
// // 		}
// //
// // 		if ts != last_sensor_insert_timestamp {
// // 			last_sensor_insert_timestamp = ts
// // 			if 2*time.Minute < last_sensor_insert_timestamp.Sub(last_calibration_time) {
// // 				should_calibrate = true
// // 				logger.Debugf("New sensors found, calibrating %v", family)
// // 				// logger.Criticalf("New sensors found, calibrating %v", family)
// // 			}
// // 		}
// // 		//.end
// //
// // 		// Compare sensor counts
// // 		current_sensor_count, err := db.NumDevicesWithLocation()
// // 		if nil != err {
// // 			logger.Error(err)
// // 			should_calibrate = true
// // 		}
// //
// // 		if last_sensor_count != current_sensor_count {
// // 			last_sensor_count = current_sensor_count
// // 			should_calibrate = true
// // 			logger.Debugf("Sensor counts don't match, calibrating %v", family)
// // 			// logger.Criticalf("Sensor counts don't match, calibrating %v", family)
// // 		}
// //
// // 		if 0 == current_sensor_count {
// // 			should_calibrate = false
// // 		}
// // 		//.end
// //
// // 		// calibrate database or pass
// // 		if should_calibrate {
// // 			// put callback function into calibration_queue
// // 			// this will schedule calibration with the
// // 			// runing calibrationWorker processes.
// // 			calibration_queue <- func() {
// // 				logger.Warnf("Calibrating %v...", family)
// // 				// if any errors occur they get swallowed
// // 				err := Calibrate(db, family, true)
// // 				if nil != err {
// // 					logger.Error(err)
// // 					return
// // 				}
// // 				logger.Infof("Calibration for %v complete", family)
// // 			}
// // 		} else {
// // 			logger.Debugf("Calibration not needed for %v", family)
// // 		}
// //
// // 		time.Sleep(60 * time.Second)
// // 	}
// // }
// //
// // // calibrationWorker reads from calibration_queue and runs AI calibration
// // func calibrationWorker() {
// // 	for clbk := range calibration_queue {
// // 		// drain queue
// // 		if 0 == len(calibration_queue) {
// // 			clbk()
// // 		}
// // 	}
// // }
