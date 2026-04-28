package domain

import "math"

func EvaluateAlarmState(
	prev AlarmState,
	value *float64,
	quality Quality,
	setpoints Setpoints,
	hysteresisPercent float64,
) (AlarmState, bool) {
	next := resolveAlarmState(value, quality, setpoints)
	if next == prev {
		return next, false
	}

	if shouldKeepPreviousByHysteresis(prev, next, value, setpoints, hysteresisPercent) {
		return prev, false
	}

	return next, true
}

func resolveAlarmState(
	value *float64,
	quality Quality,
	setpoints Setpoints,
) AlarmState {
	switch quality {
	case QualityCommLoss:
		return AlarmStateCommLoss
	case QualityOffline:
		return AlarmStateOffline
	case QualityBad:
		return AlarmStateBad
	case QualityUncertain:
		return AlarmStateUncertain
	}

	if value == nil {
		return AlarmStateBad
	}

	if setpoints.Hi == nil && setpoints.HiHi == nil && setpoints.Lo == nil && setpoints.LoLo == nil {
		return AlarmStateOK
	}

	current := *value
	if setpoints.HiHi != nil && current >= *setpoints.HiHi {
		return AlarmStateHiHi
	}
	if setpoints.Hi != nil && current >= *setpoints.Hi {
		return AlarmStateHi
	}
	if setpoints.LoLo != nil && current <= *setpoints.LoLo {
		return AlarmStateLoLo
	}
	if setpoints.Lo != nil && current <= *setpoints.Lo {
		return AlarmStateLo
	}

	return AlarmStateOK
}

func shouldKeepPreviousByHysteresis(
	prev AlarmState,
	next AlarmState,
	value *float64,
	setpoints Setpoints,
	hysteresisPercent float64,
) bool {
	if next != AlarmStateOK || value == nil {
		return false
	}

	switch prev {
	case AlarmStateHi, AlarmStateHiHi:
		margin := hysteresisMargin(setpoints, hysteresisPercent)
		highThreshold := firstNonNil(setpoints.HiHi, setpoints.Hi)
		if highThreshold == nil {
			return false
		}
		return *value > *highThreshold-margin
	case AlarmStateLo, AlarmStateLoLo:
		margin := hysteresisMargin(setpoints, hysteresisPercent)
		lowThreshold := firstNonNil(setpoints.LoLo, setpoints.Lo)
		if lowThreshold == nil {
			return false
		}
		return *value < *lowThreshold+margin
	default:
		return false
	}
}

func hysteresisMargin(setpoints Setpoints, hysteresisPercent float64) float64 {
	if hysteresisPercent <= 0 {
		hysteresisPercent = 2
	}

	if setpoints.HiHi != nil && setpoints.LoLo != nil {
		span := math.Abs(*setpoints.HiHi - *setpoints.LoLo)
		if span > 0 {
			return span * hysteresisPercent / 100
		}
	}

	if setpoints.Hi != nil && setpoints.Lo != nil {
		span := math.Abs(*setpoints.Hi - *setpoints.Lo)
		if span > 0 {
			return span * hysteresisPercent / 100
		}
	}

	return hysteresisPercent / 100
}

func firstNonNil(values ...*float64) *float64 {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}
