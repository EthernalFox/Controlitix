package snmp

import (
	"math"
	"strings"
	"testing"

	"github.com/gosnmp/gosnmp"
)

func TestDecodePDU(t *testing.T) {
	tests := []struct {
		name          string
		pdu           gosnmp.SnmpPDU
		dataType      string
		expectedValue any
		floatDelta    float64
		expectError   bool
	}{
		{
			name: "integer to int32",
			pdu: gosnmp.SnmpPDU{
				Type:  gosnmp.Integer,
				Value: int(42),
			},
			dataType:      "int32",
			expectedValue: int64(42),
		},
		{
			name: "counter32 to uint32",
			pdu: gosnmp.SnmpPDU{
				Type:  gosnmp.Counter32,
				Value: uint32(123),
			},
			dataType:      "uint32",
			expectedValue: uint64(123),
		},
		{
			name: "counter64 to float64",
			pdu: gosnmp.SnmpPDU{
				Type:  gosnmp.Counter64,
				Value: uint64(400),
			},
			dataType:      "float64",
			expectedValue: float64(400),
		},
		{
			name: "octet string to string",
			pdu: gosnmp.SnmpPDU{
				Type:  gosnmp.OctetString,
				Value: []byte("switch-core"),
			},
			dataType:      "string",
			expectedValue: "switch-core",
		},
		{
			name: "octet string numeric to float64",
			pdu: gosnmp.SnmpPDU{
				Type:  gosnmp.OctetString,
				Value: []byte("15.5"),
			},
			dataType:      "float64",
			expectedValue: float64(15.5),
			floatDelta:    1e-9,
		},
		{
			name: "object identifier to string",
			pdu: gosnmp.SnmpPDU{
				Type:  gosnmp.ObjectIdentifier,
				Value: "1.3.6.1.2.1.1.3.0",
			},
			dataType:      "string",
			expectedValue: "1.3.6.1.2.1.1.3.0",
		},
		{
			name: "opaque float to float64",
			pdu: gosnmp.SnmpPDU{
				Type:  gosnmp.OpaqueFloat,
				Value: float32(3.14),
			},
			dataType:      "float32",
			expectedValue: float64(3.14),
			floatDelta:    1e-5,
		},
		{
			name: "opaque double to float64",
			pdu: gosnmp.SnmpPDU{
				Type:  gosnmp.OpaqueDouble,
				Value: 8.5,
			},
			dataType:      "float64",
			expectedValue: float64(8.5),
			floatDelta:    1e-9,
		},
		{
			name: "no such object returns error",
			pdu: gosnmp.SnmpPDU{
				Type: gosnmp.NoSuchObject,
			},
			dataType:    "float64",
			expectError: true,
		},
		{
			name: "null returns error",
			pdu: gosnmp.SnmpPDU{
				Type: gosnmp.Null,
			},
			dataType:    "float64",
			expectError: true,
		},
		{
			name: "octet string non numeric for number returns error",
			pdu: gosnmp.SnmpPDU{
				Type:  gosnmp.OctetString,
				Value: []byte("abc"),
			},
			dataType:    "uint16",
			expectError: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			decodedValue, decodeError := DecodePDU(testCase.pdu, testCase.dataType)
			if testCase.expectError {
				if decodeError == nil {
					t.Fatal("expected decode error, got nil")
				}
				return
			}

			if decodeError != nil {
				t.Fatalf("unexpected decode error: %v", decodeError)
			}

			switch expected := testCase.expectedValue.(type) {
			case float64:
				value, ok := decodedValue.(float64)
				if !ok {
					t.Fatalf("expected float64 value, got %T", decodedValue)
				}

				delta := testCase.floatDelta
				if delta <= 0 {
					delta = 1e-9
				}
				if math.Abs(value-expected) > delta {
					t.Fatalf("unexpected float value: expected %f got %f", expected, value)
				}
			default:
				if decodedValue != expected {
					t.Fatalf("unexpected value: expected %v got %v", expected, decodedValue)
				}
			}
		})
	}
}

func TestDecodePDUIntegerToBool(t *testing.T) {
	falseValue, falseError := DecodePDU(gosnmp.SnmpPDU{
		Type:  gosnmp.Integer,
		Value: int(0),
	}, "bool")
	if falseError != nil {
		t.Fatalf("unexpected decode error: %v", falseError)
	}
	if value, ok := falseValue.(bool); !ok || value {
		t.Fatalf("expected false bool value, got %v (%T)", falseValue, falseValue)
	}

	trueValue, trueError := DecodePDU(gosnmp.SnmpPDU{
		Type:  gosnmp.Integer,
		Value: int(1),
	}, "bool")
	if trueError != nil {
		t.Fatalf("unexpected decode error: %v", trueError)
	}
	if value, ok := trueValue.(bool); !ok || !value {
		t.Fatalf("expected true bool value, got %v (%T)", trueValue, trueValue)
	}
}

func TestDecodePDUReportsNoSuchObject(t *testing.T) {
	_, decodeError := DecodePDU(gosnmp.SnmpPDU{
		Type: gosnmp.NoSuchInstance,
	}, "float64")
	if decodeError == nil {
		t.Fatal("expected decode error, got nil")
	}
	if !strings.Contains(strings.ToLower(decodeError.Error()), "no such object") {
		t.Fatalf("expected no such object error, got %v", decodeError)
	}
}
