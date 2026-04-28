package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Quality string

const (
	QualityOK           Quality = "ok"
	QualityHi           Quality = "hi"
	QualityHiHi         Quality = "hihi"
	QualityUncertain    Quality = "uncertain"
	QualityBad          Quality = "bad"
	QualityCommLoss     Quality = "comm_loss"
	QualityOffline      Quality = "offline"
	QualityAcknowledged Quality = "acknowledged"
)

func (quality Quality) ToCode() int16 {
	switch quality {
	case QualityHi:
		return 1
	case QualityHiHi:
		return 2
	case QualityUncertain:
		return 3
	case QualityBad:
		return 4
	case QualityCommLoss:
		return 5
	case QualityOffline:
		return 6
	case QualityAcknowledged:
		return 7
	default:
		return 0
	}
}

func QualityFromCode(code int16) Quality {
	switch code {
	case 1:
		return QualityHi
	case 2:
		return QualityHiHi
	case 3:
		return QualityUncertain
	case 4:
		return QualityBad
	case 5:
		return QualityCommLoss
	case 6:
		return QualityOffline
	case 7:
		return QualityAcknowledged
	default:
		return QualityOK
	}
}

func ParseQuality(value string) (Quality, error) {
	switch Quality(value) {
	case QualityOK,
		QualityHi,
		QualityHiHi,
		QualityUncertain,
		QualityBad,
		QualityCommLoss,
		QualityOffline,
		QualityAcknowledged:
		return Quality(value), nil
	default:
		return "", errors.New("invalid quality")
	}
}

type Aggregator string

const (
	AggregatorLast Aggregator = "last"
	AggregatorAvg  Aggregator = "avg"
	AggregatorMin  Aggregator = "min"
	AggregatorMax  Aggregator = "max"
)

func ParseAggregator(value string) (Aggregator, error) {
	switch Aggregator(value) {
	case AggregatorLast, AggregatorAvg, AggregatorMin, AggregatorMax:
		return Aggregator(value), nil
	default:
		return "", errors.New("invalid aggregator")
	}
}

type TrendSource string

const (
	TrendSourceRaw   TrendSource = "raw"
	TrendSourceAgg1m TrendSource = "agg_1m"
)

type TrendQuery struct {
	TagID       uuid.UUID
	From        time.Time
	To          time.Time
	Step        time.Duration
	Aggregator  Aggregator
	Limit       int
}

type TrendBatchQuery struct {
	TagIDs      []uuid.UUID
	From        time.Time
	To          time.Time
	Step        time.Duration
	Aggregator  Aggregator
	Limit       int
}

type TrendPoint struct {
	Timestamp time.Time
	Value     *float64
	Quality   Quality
}

type Unit struct {
	ID       int64
	Name     string
	Symbol   string
	Category string
}

type DataType struct {
	ID   int64
	Name string
}

type TagMeta struct {
	TagID      uuid.UUID
	TagName    string
	DeviceID   uuid.UUID
	DeviceName string
	Unit       Unit
	DataType   DataType
}

type TrendSeries struct {
	TagMeta
	From       time.Time
	To         time.Time
	Step       time.Duration
	Aggregator Aggregator
	Source     TrendSource
	Points     []TrendPoint
}

type TrendProblem struct {
	Type   string
	Title  string
	Status int
	Detail string
}

type TrendBatchError struct {
	TagID   uuid.UUID
	Problem TrendProblem
}

type TrendBatchResult struct {
	Series []TrendSeries
	Errors []TrendBatchError
}

type IngestRecord struct {
	TagID      uuid.UUID
	Timestamp  time.Time
	Value      *float64
	Quality    Quality
}
