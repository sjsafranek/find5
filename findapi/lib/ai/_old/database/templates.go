package database

import (
	"time"

	"github.com/schollz/find4/server/main/src/models"
)

// SENSOR_SQL is the sql json template for a sensor object.
var SENSOR_SQL = `
	'{'||
		'"timestamp": ' ||  timestamp ||','||
		'"deviceid": "' ||  deviceid ||'",'||
		'"locationid": "' ||  locationid ||'",'||
		'"create_at": "' ||  create_at ||'",'||
		'"update_at": "' ||  update_at ||'",'||
		'"sensor_type": "' ||  sensor_type ||'",'||
		'"sensor": ' ||  sensor
	|| '}'
`

// LOCATION_PREDICTION_SQL is the sql json template for
// a location_prediction object.
var LOCATION_PREDICTION_SQL = `
	'{'||
		'"timestamp": ' ||  timestamp ||','||
		'"location": "' ||  locationid ||'",'||
		'"create_at": "' ||  create_at ||'",'||
		'"update_at": "' ||  update_at ||'",'||
		'"probability": ' ||  probability
	|| '}'
`

// CALIBRATION_SQL is the sql json template for
// calibration objects
var CALIBRATION_SQL = `
	'{'||
		'"id": '|| id ||','||
		'"probability_means": ' ||
			CASE
			    WHEN probability_means IS NULL THEN 'null'
			    WHEN ''=probability_means THEN 'null'
			    ELSE probability_means
			END
		||','||
		'"probabilities_of_best_guess": ' || probabilities_of_best_guess ||','||
		'"percent_correct": ' || percent_correct ||','||
		'"accuracy_breakdown": ' || accuracy_breakdown ||','||
		'"prediction_analysis": ' || prediction_analysis ||','||
		'"algorithm_efficacy": ' || algorithm_efficacy ||','||
		'"calibration_time": "' || strftime('%Y-%m-%dT%H:%M:%SZ', calibration_time) ||'",'||
		'"create_at": "' || strftime('%Y-%m-%dT%H:%M:%SZ', create_at) ||'",'||
		'"update_at": "' || strftime('%Y-%m-%dT%H:%M:%SZ', update_at) ||'"'
	|| '}'
`

type CalibrationModel struct {
	Id                       int                                      `json:"id"`
	Probability_means        []float64                                `json:"probability_means"`
	ProbabilitiesOfBestGuess []float64                                `json:"probabilities_of_best_guess"`
	PercentCorrect           float64                                  `json:"percent_correct"`
	AccuracyBreakdown        map[string]float64                       `json:"accuracy_breakdown"`
	PredictionAnalysis       map[string]map[string]map[string]int     `json:"prediction_analysis"`
	AlgorithmEfficacy        map[string]map[string]models.BinaryStats `json:"algorithm_efficacy"`
	CalibrationTime          time.Time                                `json:"calibration_time"`
	CreateAt                 time.Time                                `json:"create_at"`
	UpdateAt                 time.Time                                `json:"update_at"`
}
