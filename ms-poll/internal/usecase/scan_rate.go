package usecase

import (
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

const (
	defaultScanRate     = 1 * time.Second
	temperatureScanRate = 60 * time.Second
)

var slowTemperatureUnits = map[string]struct{}{
	"°c": {},
	"°f": {},
	"k":  {},
}

func ScanRateFor(tag *domain.TagSnapshot) time.Duration {
	if tag == nil || tag.UnitSymbol == nil {
		return defaultScanRate
	}

	normalizedUnit := strings.ToLower(strings.TrimSpace(*tag.UnitSymbol))
	if _, isSlowUnit := slowTemperatureUnits[normalizedUnit]; isSlowUnit {
		return temperatureScanRate
	}

	return defaultScanRate
}
