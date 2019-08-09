package ai

import (
	"time"

	"github.com/sjsafranek/find5/lib/ai/models"
)

type CalibrationModel struct {
	Id                       int                                      `json:"id"`
	ProbabilityMeans         []float64                                `json:"probability_means"`
	ProbabilitiesOfBestGuess []float64                                `json:"probabilities_of_best_guess"`
	PercentCorrect           float64                                  `json:"percent_correct"`
	AccuracyBreakdown        map[string]float64                       `json:"accuracy_breakdown"`
	PredictionAnalysis       map[string]map[string]map[string]int     `json:"prediction_analysis"`
	AlgorithmEfficacy        map[string]map[string]models.BinaryStats `json:"algorithm_efficacy"`
	CalibrationTime          time.Time                                `json:"calibration_time"`
	CreateAt                 time.Time                                `json:"create_at"`
	UpdateAt                 time.Time                                `json:"update_at"`
}

type AnalysisResponse struct {
	Data    models.LocationAnalysis `json:"analysis"`
	Message string                  `json:"message"`
	Success bool                    `json:"success"`
}
