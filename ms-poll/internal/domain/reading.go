package domain

import "time"

type Quality string

const (
	QualityGood      Quality = "good"
	QualityBad       Quality = "bad"
	QualityUncertain Quality = "uncertain"
)

type Reading struct {
	TagID    string
	DeviceID string
	Value    any
	RawValue any
	Quality  Quality
	Error    string
	ReadAt   time.Time
}
