package domain

import "testing"

func floatPtr(value float64) *float64 {
	return &value
}

func TestEvaluateAlarmStateTransitions(t *testing.T) {
	setpoints := Setpoints{
		LoLo: floatPtr(10),
		Lo:   floatPtr(20),
		Hi:   floatPtr(80),
		HiHi: floatPtr(95),
	}

	testCases := []struct {
		name      string
		prev      AlarmState
		value     *float64
		quality   Quality
		expected  AlarmState
		changed   bool
	}{
		{
			name:     "ok to hi",
			prev:     AlarmStateOK,
			value:    floatPtr(81),
			quality:  QualityOK,
			expected: AlarmStateHi,
			changed:  true,
		},
		{
			name:     "hi to hihi",
			prev:     AlarmStateHi,
			value:    floatPtr(96),
			quality:  QualityOK,
			expected: AlarmStateHiHi,
			changed:  true,
		},
		{
			name:     "hihi keeps by hysteresis",
			prev:     AlarmStateHiHi,
			value:    floatPtr(94.0),
			quality:  QualityOK,
			expected: AlarmStateHiHi,
			changed:  false,
		},
		{
			name:     "hi clears to ok",
			prev:     AlarmStateHi,
			value:    floatPtr(70),
			quality:  QualityOK,
			expected: AlarmStateOK,
			changed:  true,
		},
		{
			name:     "any to comm loss",
			prev:     AlarmStateHi,
			value:    floatPtr(70),
			quality:  QualityCommLoss,
			expected: AlarmStateCommLoss,
			changed:  true,
		},
		{
			name:     "comm loss to ok",
			prev:     AlarmStateCommLoss,
			value:    floatPtr(50),
			quality:  QualityOK,
			expected: AlarmStateOK,
			changed:  true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			next, changed := EvaluateAlarmState(
				testCase.prev,
				testCase.value,
				testCase.quality,
				setpoints,
				2,
			)

			if next != testCase.expected {
				t.Fatalf("expected state %s, got %s", testCase.expected.String(), next.String())
			}
			if changed != testCase.changed {
				t.Fatalf("expected changed=%v, got %v", testCase.changed, changed)
			}
		})
	}
}
