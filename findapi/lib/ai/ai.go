package ai

import (
	"bytes"
	"compress/gzip"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"github.com/sjsafranek/find5/findapi/lib/ai/learning/nb1"
	"github.com/sjsafranek/find5/findapi/lib/ai/learning/nb2"
	"github.com/sjsafranek/find5/findapi/lib/ai/models"
	"github.com/sjsafranek/find5/findapi/lib/database"
	"github.com/sjsafranek/ligneous"
	"github.com/sjsafranek/pool"
)

var (
	logger = ligneous.AddLogger("ai", "trace", "./log/find5")
)

func SetLoggingDirectory(directory string) {
	logger = ligneous.AddLogger("ai", "trace", directory)
}

func New(aiConnStr, redisAddr string) *AI {

	factory := func() (net.Conn, error) { return net.Dial("tcp", aiConnStr) }
	var aiPool pool.Pool
	for {
		connPool, err := pool.NewChannelPool(4, 10, factory)
		if nil != err {
			logger.Warn("Unable to communicate with AI server")
			time.Sleep(5 * time.Second)
			continue
		}
		logger.Info("Connected to AI server")
		aiPool = connPool
		break
	}

	ai := AI{
		// aiConnStr: aiConnStr,
		aiPool: aiPool,
		redis: &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", redisAddr) },
		},
	}

	go func() {
		for {
			time.Sleep(10 * time.Second)
			ai.guard.RLock()
			if 0 != ai.pending {
				logger.Debugf("%v pending AI requests", ai.pending)
			}
			ai.guard.RUnlock()
		}
	}()

	return &ai

}

type AI struct {
	// aiConnStr string
	redis   *redis.Pool
	aiPool  pool.Pool
	pending int
	guard   sync.RWMutex
}

func (self *AI) Set(key string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	conn := self.redis.Get()
	defer conn.Close()
	_, err = redis.Bytes(conn.Do("SET", key, data))
	return err
}

func (self *AI) Get(key string, v interface{}) error {
	conn := self.redis.Get()
	defer conn.Close()
	data, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &v)
}

// HACK
func (self *AI) convertMeasurementsToSensorData(locationMeasurements []*database.LocationMeasurements, family string) []models.SensorData {

	datas := []models.SensorData{}

	locationTimeBucketData := make(map[string]models.SensorData)

	for _, location := range locationMeasurements {
		for _, sensor := range location.SensorMeasurements {
			for _, bucket := range sensor.BucketMeasurements {

				hsh := fmt.Sprintf("%v-%v", location.LocationId, bucket.BucketId)

				// if location already in map
				if _, ok := locationTimeBucketData[hsh]; ok {
					sd := models.SensorData{
						Family:    family,
						Device:    sensor.DeviceId,
						Location:  location.LocationId,
						Timestamp: bucket.BucketId,
						Sensors:   make(map[string]map[string]interface{}),
					}
					sd.Sensors[sensor.SensorId] = make(map[string]interface{})
					for _, measurement := range bucket.Measurements {
						locationTimeBucketData[hsh].Sensors[sensor.SensorId][measurement.Key] = measurement.Value
					}
					continue
				}

				// create new locdation bucket
				sd := models.SensorData{
					Family:    family,
					Device:    sensor.DeviceId,
					Location:  location.LocationId,
					Timestamp: bucket.BucketId,
					Sensors:   make(map[string]map[string]interface{}),
				}
				sd.Sensors[sensor.SensorId] = make(map[string]interface{})
				for _, measurement := range bucket.Measurements {
					sd.Sensors[sensor.SensorId][measurement.Key] = measurement.Value
				}
				locationTimeBucketData[hsh] = sd

			}
		}
	}

	for i := range locationTimeBucketData {
		datas = append(datas, locationTimeBucketData[i])
	}

	return datas
}

//.end

func (self *AI) Calibrate(locationMeasurements []*database.LocationMeasurements, family string, crossValidation ...bool) error {
	datas := self.convertMeasurementsToSensorData(locationMeasurements, family)

	datasLearn, datasTest, err := self.splitDataForLearning(datas, crossValidation...)
	if err != nil {
		return err
	}

	// do the Golang naive bayes fitting
	nb := nb1.New()
	logger.Debugf("naive bayes1 fitting")
	errFit := nb.Fit(datasLearn)
	if errFit != nil {
		logger.Error(errFit)
	} else {
		_ = self.Set(fmt.Sprintf("NB1-%v", family), nb.Data)
	}

	// do the Golang naive bayes2 fitting
	nbFit2 := nb2.New()
	logger.Debugf("naive bayes2 fitting")
	errFit = nbFit2.Fit(datasLearn)
	if errFit != nil {
		logger.Error(errFit)
	} else {
		_ = self.Set(fmt.Sprintf("NB2-%v", family), nbFit2.Data)
	}

	// do the python learning
	err = self.learnFromData(family, datasLearn)
	if err != nil {
		return err
	}

	if len(crossValidation) > 0 && crossValidation[0] {
		go self.findBestAlgorithm(datasTest, family)
	}

	return err
}

func (self *AI) splitDataForLearning(datas []models.SensorData, crossValidation ...bool) ([]models.SensorData, []models.SensorData, error) {
	datasLearn := []models.SensorData{}
	datasTest := []models.SensorData{}

	if len(datas) < 2 {
		return datasLearn, datasTest, errors.New("not enough data")
	}
	// for cross validation only
	if len(crossValidation) > 0 && crossValidation[0] {
		// randomize data order
		for i := range datas {
			j := rand.Intn(i + 1)
			datas[i], datas[j] = datas[j], datas[i]
		}
		if len(datas) > 1000 {
			datas = datas[:1000]
		}

		// triage into different locations
		dataLocations := make(map[string][]int)
		for i := range datas {
			if _, ok := dataLocations[datas[i].Location]; !ok {
				dataLocations[datas[i].Location] = []int{}
			}
			dataLocations[datas[i].Location] = append(dataLocations[datas[i].Location], i)
		}

		// for each location, make test set and learn set
		datasTest = make([]models.SensorData, len(datas))
		datasTestI := 0
		datasLearn = make([]models.SensorData, len(datas))
		datasLearnI := 0
		for loc := range dataLocations {
			splitI := 1
			numDataPoints := len(dataLocations[loc])
			if numDataPoints < 2 {
				continue
			} else if numDataPoints < 10 {
				splitI = numDataPoints / 2 // 50% split
			} else {
				splitI = numDataPoints * 7 / 10 // 70:30 split
			}
			for i, s := range dataLocations[loc] {
				if i < splitI {
					// used for learning
					datasLearn[datasLearnI] = datas[s]
					datasLearnI++
				} else {
					datasTest[datasTestI] = datas[s]
					datasTestI++
				}
			}
			logger.Debugf("splitting %s data for cross validation (%d -> %d)", loc, numDataPoints, splitI)
		}

		datasLearn = datasLearn[:datasLearnI]
		datasTest = datasTest[:datasTestI]
		logger.Debugf("[%s]  learning: %d, testing: %d", datas[0].Family, len(datas), len(datasTest))
	}
	return datasLearn, datasTest, nil
}

func (self *AI) learnFromData(family string, datas []models.SensorData) error {

	b64Data, err := formatFilePayload(datas)
	if err != nil {
		return err
	}

	body, err := self.aiSendAndRecieve(fmt.Sprintf(`{"method":"learn","data":{"family":"%v","file_data":"%v"}}`, family, b64Data))
	if nil != err {
		return errors.Wrap(err, "problem sending message to ai server")
	}

	var target AnalysisResponse
	err = json.Unmarshal([]byte(body), &target)
	if err != nil {
		return errors.Wrap(err, "problem decoding response")
	}

	if target.Success {
		logger.Debugf("success: %s", target.Message)
	} else {
		logger.Debugf("failure: %s", target.Message)
		return errors.New("failed in AI server: " + target.Message)
	}
	return nil
}

func (self *AI) findBestAlgorithm(datas []models.SensorData, family string) (algorithmEfficacy map[string]map[string]models.BinaryStats, err error) {
	if len(datas) == 0 {
		err = errors.New("no data specified")
		return
	}
	predictionAnalysis := make(map[string]map[string]map[string]int)
	logger.Debugf("[%s] finding best algorithm for %d data", datas[0].Family, len(datas))

	t := time.Now()
	type Job struct {
		data models.SensorData
		i    int
	}
	type Result struct {
		data models.LocationAnalysis
		i    int
	}
	jobs := make(chan Job, len(datas))
	results := make(chan Result, len(datas))
	workers := 9
	for w := 0; w < workers; w++ {
		go func(id int, jobs <-chan Job, results chan<- Result) {
			for job := range jobs {
				aidata, err := self.AnalyzeSensorData(job.data, family)
				if err != nil {
					logger.Warnf("%s: %+v", err.Error(), job.data)
				}
				results <- Result{data: aidata, i: job.i}
			}
		}(w, jobs, results)
	}
	for i, data := range datas {
		jobs <- Job{data: data, i: i}
	}
	close(jobs)
	aidatas := make([]models.LocationAnalysis, len(datas))
	for i := 0; i < len(datas); i++ {
		result := <-results
		aidatas[result.i] = result.data
	}
	logger.Infof("[%s] analyzed %d data in %s", datas[0].Family, len(datas), time.Since(t))

	for i, aidata := range aidatas {
		for _, prediction := range aidata.Predictions {
			if _, ok := predictionAnalysis[prediction.Name]; !ok {
				predictionAnalysis[prediction.Name] = make(map[string]map[string]int)
				for trueLoc := range aidata.LocationNames {
					predictionAnalysis[prediction.Name][aidata.LocationNames[trueLoc]] = make(map[string]int)
					for guessLoc := range aidata.LocationNames {
						predictionAnalysis[prediction.Name][aidata.LocationNames[trueLoc]][aidata.LocationNames[guessLoc]] = 0
					}
				}
			}
			correctLocation := datas[i].Location
			if len(prediction.Locations) == 0 {
				logger.Warn("prediction.Locations is empty!")
				continue
			}
			if len(aidata.LocationNames) == 0 {
				err = errors.New("no location names")
				logger.Error(err)
				return
			}

			guessedLocation := aidata.LocationNames[prediction.Locations[0]]
			predictionAnalysis[prediction.Name][correctLocation][guessedLocation]++
		}
	}

	// normalize prediction analysis
	// initialize location totals
	locationTotals := make(map[string]int)
	for _, data := range datas {
		if _, ok := locationTotals[data.Location]; !ok {
			locationTotals[data.Location] = 0
		}
		locationTotals[data.Location]++
	}
	logger.Debugf("locationTotals: %+v", locationTotals)
	algorithmEfficacy = make(map[string]map[string]models.BinaryStats)
	for alg := range predictionAnalysis {
		if _, ok := algorithmEfficacy[alg]; !ok {
			algorithmEfficacy[alg] = make(map[string]models.BinaryStats)
		}
		for correctLocation := range predictionAnalysis[alg] {
			// calculate true/false positives/negatives
			tp := 0
			fp := 0
			tn := 0
			fn := 0
			for guessedLocation := range predictionAnalysis[alg][correctLocation] {
				count := predictionAnalysis[alg][correctLocation][guessedLocation]
				if guessedLocation == correctLocation {
					tp += count
				} else if guessedLocation != correctLocation {
					fn += count
				}
			}
			for otherCorrectLocation := range predictionAnalysis[alg] {
				if otherCorrectLocation == correctLocation {
					continue
				}
				for guessedLocation := range predictionAnalysis[alg] {
					count := predictionAnalysis[alg][otherCorrectLocation][guessedLocation]
					if guessedLocation == correctLocation {
						fp += count
					} else if guessedLocation != correctLocation {
						tn += count
					}
				}
			}
			algorithmEfficacy[alg][correctLocation] = models.NewBinaryStats(tp, fp, tn, fn)
		}
	}

	correct := 0
	ProbabilitiesOfBestGuess := make([]float64, len(aidatas))
	accuracyBreakdown := make(map[string]float64)
	accuracyBreakdownTotal := make(map[string]float64)
	for i := range aidatas {
		if _, ok := accuracyBreakdownTotal[datas[i].Location]; !ok {
			accuracyBreakdownTotal[datas[i].Location] = 0
			accuracyBreakdown[datas[i].Location] = 0
		}
		accuracyBreakdownTotal[datas[i].Location]++
		bestGuess := determineBestGuess(aidatas[i], algorithmEfficacy)
		if len(bestGuess) == 0 {
			continue
		}
		if bestGuess[0].Location == datas[i].Location {
			accuracyBreakdown[datas[i].Location]++
			correct++
			ProbabilitiesOfBestGuess[i] = bestGuess[0].Probability
		} else {
			ProbabilitiesOfBestGuess[i] = -1 * bestGuess[0].Probability
		}
	}
	logger.Infof("[%s] total correct: %d/%d", datas[0].Family, correct, len(aidatas))

	goodProbs := make([]float64, len(ProbabilitiesOfBestGuess))
	i := 0
	for _, v := range ProbabilitiesOfBestGuess {
		if v > 0 {
			goodProbs[i] = v
			i++
		}
	}
	goodProbs = goodProbs[:i]
	goodMean := average(goodProbs)
	goodSD := stdDev(goodProbs, goodMean)

	badProbs := make([]float64, len(ProbabilitiesOfBestGuess))
	i = 0
	for _, v := range ProbabilitiesOfBestGuess {
		if v < 0 {
			badProbs[i] = -1 * v
			i++
		}
	}
	badProbs = badProbs[:i]
	badMean := average(badProbs)
	badSD := stdDev(badProbs, badMean)

	for loc := range accuracyBreakdown {
		accuracyBreakdown[loc] = accuracyBreakdown[loc] / accuracyBreakdownTotal[loc]
		logger.Infof("[%s] %s accuracy: %2.0f%%", datas[0].Family, loc, accuracyBreakdown[loc]*100)
	}

	calibrationModel := CalibrationModel{
		ProbabilityMeans:         []float64{goodMean, goodSD, badMean, badSD},
		ProbabilitiesOfBestGuess: ProbabilitiesOfBestGuess,
		PercentCorrect:           float64(correct) / float64(len(datas)),
		AccuracyBreakdown:        accuracyBreakdown,
		PredictionAnalysis:       predictionAnalysis,
		AlgorithmEfficacy:        algorithmEfficacy,
	}
	self.Set(fmt.Sprintf("calibration-%v", family), calibrationModel)

	return
}

func formatFilePayload(datas []models.SensorData) (string, error) {
	if len(datas) == 0 {
		return "", errors.New("data is empty")
	}

	// determine all possible columns
	sensorColumns := make(map[string]int)
	columnCount := 1
	for _, data := range datas {
		for sensorType := range data.Sensors {
			for sensorName := range data.Sensors[sensorType] {
				name := fmt.Sprintf("%s-%s", sensorType, sensorName)
				if _, ok := sensorColumns[name]; !ok {
					sensorColumns[name] = columnCount
					columnCount++
				}
			}
		}
	}

	// get column names
	columns := make([]string, columnCount)
	columns[0] = "location"
	for column := range sensorColumns {
		columns[sensorColumns[column]] = column
	}

	csv_data := strings.Join(columns, ",") + "\n"

	for _, data := range datas {
		columns = make([]string, columnCount)
		columns[0] = data.Location
		for sensorType := range data.Sensors {
			for sensorName := range data.Sensors[sensorType] {
				cId := fmt.Sprintf("%s-%s", sensorType, sensorName)
				columns[sensorColumns[cId]] = fmt.Sprintf("%3.9f", data.Sensors[sensorType][sensorName])
			}
		}
		line := strings.Join(columns, ",") + "\n"
		csv_data += line
	}

	// compress
	var buff bytes.Buffer
	gz := gzip.NewWriter(&buff)
	gz.Write([]byte(csv_data))
	gz.Close()
	payload := b64.StdEncoding.EncodeToString(buff.Bytes())

	return payload, nil
}
