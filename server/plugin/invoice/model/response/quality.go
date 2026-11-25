package response

import "time"

type QualityDashboard struct {
	TotalRecognitions      int64   `json:"totalRecognitions"`
	SuccessfulRecognitions int64   `json:"successfulRecognitions"`
	FailedRecognitions     int64   `json:"failedRecognitions"`
	SuccessRate            float64 `json:"successRate"`
	FailureRate            float64 `json:"failureRate"`
	AverageDurationMS      int64   `json:"averageDurationMs"`
	AverageAttempts        float64 `json:"averageAttempts"`
	MultimodalFallbackRate float64 `json:"multimodalFallbackRate"`
	ReviewedInvoices       int64   `json:"reviewedInvoices"`
	CorrectedFields        int64   `json:"correctedFields"`
	LegacyWithoutFieldData int64   `json:"legacyWithoutFieldData"`
	EstimatedCostMicros    int64   `json:"estimatedCostMicros"`
}

type ProviderMetric struct {
	Provider          string  `json:"provider"`
	Model             string  `json:"model"`
	FileType          string  `json:"fileType"`
	Total             int64   `json:"total"`
	Success           int64   `json:"success"`
	Failed            int64   `json:"failed"`
	SuccessRate       float64 `json:"successRate"`
	AverageConfidence float64 `json:"averageConfidence"`
	AverageDurationMS int64   `json:"averageDurationMs"`
	AverageAttempts   float64 `json:"averageAttempts"`
	CorrectedFields   int64   `json:"correctedFields"`
}

type FieldMetric struct {
	FieldName         string  `json:"fieldName"`
	Label             string  `json:"label"`
	Reviewed          int64   `json:"reviewed"`
	Modified          int64   `json:"modified"`
	ModificationRate  float64 `json:"modificationRate"`
	AccuracyRate      float64 `json:"accuracyRate"`
	AverageConfidence float64 `json:"averageConfidence"`
}

type QualityFailure struct {
	InvoiceID   uint      `json:"invoiceId"`
	FileName    string    `json:"fileName"`
	FileType    string    `json:"fileType"`
	Provider    string    `json:"provider"`
	Model       string    `json:"model"`
	Attempts    int       `json:"attempts"`
	MaxAttempts int       `json:"maxAttempts"`
	Error       string    `json:"error"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type ClassificationMetric struct {
	Suggested      int64   `json:"suggested"`
	Accepted       int64   `json:"accepted"`
	Overridden     int64   `json:"overridden"`
	Pending        int64   `json:"pending"`
	AcceptanceRate float64 `json:"acceptanceRate"`
	OverrideRate   float64 `json:"overrideRate"`
}
