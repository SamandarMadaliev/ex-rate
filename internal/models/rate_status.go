package models

type RateStatus string

const (
	RateStatusPending   RateStatus = "pending"
	RateStatusCompleted RateStatus = "completed"
	RateStatusFailed    RateStatus = "failed"
)
