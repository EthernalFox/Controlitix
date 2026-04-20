package usecase

import (
	"testing"
	"time"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

func TestScanRateForCelsiusIsSlow(t *testing.T) {
	tag := &domain.TagSnapshot{
		UnitSymbol: pointerToString("°C"),
	}

	scanRate := ScanRateFor(tag)
	if scanRate != 60*time.Second {
		t.Fatalf("expected 60s for °C, got %s", scanRate)
	}
}

func TestScanRateForVoltIsDefault(t *testing.T) {
	tag := &domain.TagSnapshot{
		UnitSymbol: pointerToString("V"),
	}

	scanRate := ScanRateFor(tag)
	if scanRate != 1*time.Second {
		t.Fatalf("expected 1s for V, got %s", scanRate)
	}
}

func TestScanRateForNilUnitIsDefault(t *testing.T) {
	tag := &domain.TagSnapshot{}

	scanRate := ScanRateFor(tag)
	if scanRate != 1*time.Second {
		t.Fatalf("expected 1s for nil unit, got %s", scanRate)
	}
}
