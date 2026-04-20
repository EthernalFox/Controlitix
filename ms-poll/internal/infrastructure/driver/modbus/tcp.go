package modbus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
	goburrowmodbus "github.com/goburrow/modbus"
)

const (
	defaultTCPPort      = 502
	defaultUnitID       = 1
	defaultTimeoutMs    = 1000
	defaultRTUBaudRate  = 9600
	defaultRTUDataBits  = 8
	defaultRTUParity    = "N"
	defaultRTUStopBits  = 1
)

type TCPDriverFactory struct{}

func NewTCPDriverFactory() *TCPDriverFactory {
	return &TCPDriverFactory{}
}

func (factory *TCPDriverFactory) SupportedType() string {
	return "modbus_tcp"
}

func (factory *TCPDriverFactory) Build(
	device *domain.DeviceSnapshot,
) (domain.Driver, error) {
	if device == nil {
		return nil, errors.New("device snapshot is nil")
	}

	settings, parseError := parseTCPSettings(device.Settings)
	if parseError != nil {
		return nil, fmt.Errorf("parse tcp settings: %w", parseError)
	}

	handler := goburrowmodbus.NewTCPClientHandler(
		fmt.Sprintf("%s:%d", settings.Host, settings.Port),
	)
	handler.Timeout = time.Duration(settings.TimeoutMs) * time.Millisecond
	handler.SlaveId = settings.UnitID

	connectError := handler.Connect()
	if connectError != nil {
		return nil, fmt.Errorf("connect tcp handler: %w", connectError)
	}

	return &tcpDriver{
		handler: handler,
		client:  goburrowmodbus.NewClient(handler),
	}, nil
}

type tcpDriver struct {
	handler *goburrowmodbus.TCPClientHandler
	client  modbusClient
}

func (driver *tcpDriver) Name() string {
	return "modbus_tcp"
}

func (driver *tcpDriver) Read(
	ctx context.Context,
	tags []*domain.TagSnapshot,
) []domain.Reading {
	return readWithClient(ctx, tags, driver.client)
}

func (driver *tcpDriver) HealthCheck(ctx context.Context) error {
	if ctxError := ctx.Err(); ctxError != nil {
		return ctxError
	}

	_, holdingError := driver.client.ReadHoldingRegisters(0, 1)
	if holdingError == nil {
		return nil
	}

	_, coilsError := driver.client.ReadCoils(0, 1)
	if coilsError == nil {
		return nil
	}

	return fmt.Errorf(
		"tcp health check failed: %w",
		errors.Join(holdingError, coilsError),
	)
}

func (driver *tcpDriver) Close() error {
	if driver.handler == nil {
		return nil
	}

	closeError := driver.handler.Close()
	if closeError != nil {
		return fmt.Errorf("close tcp handler: %w", closeError)
	}

	return nil
}

type tcpSettings struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	UnitID    byte   `json:"unit_id"`
	TimeoutMs int    `json:"timeout_ms"`
}

func parseTCPSettings(rawSettings json.RawMessage) (tcpSettings, error) {
	settings := tcpSettings{
		Port:      defaultTCPPort,
		UnitID:    defaultUnitID,
		TimeoutMs: defaultTimeoutMs,
	}

	if len(rawSettings) > 0 {
		unmarshalError := json.Unmarshal(rawSettings, &settings)
		if unmarshalError != nil {
			return tcpSettings{}, fmt.Errorf(
				"unmarshal tcp settings: %w",
				unmarshalError,
			)
		}
	}

	settings.Host = trimSpace(settings.Host)
	if settings.Host == "" {
		return tcpSettings{}, errors.New("host is required for modbus_tcp")
	}
	if settings.Port <= 0 {
		settings.Port = defaultTCPPort
	}
	if settings.TimeoutMs <= 0 {
		settings.TimeoutMs = defaultTimeoutMs
	}
	if settings.UnitID == 0 {
		settings.UnitID = defaultUnitID
	}

	return settings, nil
}

type modbusClient interface {
	ReadHoldingRegisters(address, quantity uint16) ([]byte, error)
	ReadInputRegisters(address, quantity uint16) ([]byte, error)
	ReadCoils(address, quantity uint16) ([]byte, error)
	ReadDiscreteInputs(address, quantity uint16) ([]byte, error)
}

func readWithClient(
	ctx context.Context,
	tags []*domain.TagSnapshot,
	client modbusClient,
) []domain.Reading {
	if len(tags) == 0 {
		return nil
	}

	now := time.Now().UTC()
	result := make([]domain.Reading, len(tags))
	filled := make([]bool, len(tags))
	indexByTag := make(map[*domain.TagSnapshot]int, len(tags))

	for index, tag := range tags {
		indexByTag[tag] = index
	}

	requests, buildError := BuildBatches(tags)
	if buildError != nil {
		for index, tag := range tags {
			result[index] = newBadReading(tag, buildError, now)
		}
		return result
	}

	for _, request := range requests {
		if ctxError := ctx.Err(); ctxError != nil {
			markRequestAsBad(
				result,
				filled,
				indexByTag,
				request,
				ctxError,
			)
			continue
		}

		responsePayload, readError := executeReadRequest(client, request)
		if readError != nil {
			markRequestAsBad(
				result,
				filled,
				indexByTag,
				request,
				readError,
			)
			continue
		}

		for _, slot := range request.Tags {
			if slot == nil || slot.Tag == nil {
				continue
			}

			tagIndex, exists := indexByTag[slot.Tag]
			if !exists {
				continue
			}

			decodedValue, decodeError := decodeSlot(request, slot, responsePayload)
			if decodeError != nil {
				result[tagIndex] = newBadReading(slot.Tag, decodeError, time.Now().UTC())
				filled[tagIndex] = true
				continue
			}

			result[tagIndex] = domain.Reading{
				TagID:    slot.Tag.ID,
				DeviceID: slot.Tag.DeviceID,
				Value:    decodedValue,
				RawValue: decodedValue,
				Quality:  domain.QualityGood,
				ReadAt:   time.Now().UTC(),
			}
			filled[tagIndex] = true
		}
	}

	for index, tag := range tags {
		if filled[index] {
			continue
		}

		result[index] = newBadReading(
			tag,
			errors.New("reading is not available"),
			time.Now().UTC(),
		)
	}

	return result
}

func executeReadRequest(
	client modbusClient,
	request ReadRequest,
) ([]byte, error) {
	switch request.RegisterType {
	case RegHolding:
		return client.ReadHoldingRegisters(
			request.StartOffset,
			request.Quantity,
		)
	case RegInput:
		return client.ReadInputRegisters(
			request.StartOffset,
			request.Quantity,
		)
	case RegCoil:
		return client.ReadCoils(
			request.StartOffset,
			request.Quantity,
		)
	case RegDiscrete:
		return client.ReadDiscreteInputs(
			request.StartOffset,
			request.Quantity,
		)
	default:
		return nil, fmt.Errorf(
			"unsupported register_type: %s",
			request.RegisterType,
		)
	}
}

func markRequestAsBad(
	result []domain.Reading,
	filled []bool,
	indexByTag map[*domain.TagSnapshot]int,
	request ReadRequest,
	readError error,
) {
	readAt := time.Now().UTC()
	for _, slot := range request.Tags {
		if slot == nil || slot.Tag == nil {
			continue
		}

		tagIndex, exists := indexByTag[slot.Tag]
		if !exists {
			continue
		}

		result[tagIndex] = newBadReading(slot.Tag, readError, readAt)
		filled[tagIndex] = true
	}
}

func decodeSlot(
	request ReadRequest,
	slot *tagSlot,
	payload []byte,
) (any, error) {
	switch request.RegisterType {
	case RegCoil, RegDiscrete:
		return DecodeBits(payload, int(slot.Offset)), nil
	case RegHolding, RegInput:
		startByte := int(slot.Offset) * 2
		endByte := startByte + (int(slot.Length) * 2)
		if endByte > len(payload) {
			return nil, fmt.Errorf(
				"response payload is too short for tag %s",
				slot.Tag.ID,
			)
		}
		return DecodeRegisters(payload[startByte:endByte], slot.Tag.DataType)
	default:
		return nil, fmt.Errorf(
			"unsupported register_type: %s",
			request.RegisterType,
		)
	}
}

func newBadReading(
	tag *domain.TagSnapshot,
	readError error,
	readAt time.Time,
) domain.Reading {
	if tag == nil {
		return domain.Reading{
			Quality: domain.QualityBad,
			Error:   "tag is nil",
			ReadAt:  readAt,
		}
	}

	errorMessage := "read failed"
	if readError != nil {
		errorMessage = readError.Error()
	}

	return domain.Reading{
		TagID:    tag.ID,
		DeviceID: tag.DeviceID,
		Quality:  domain.QualityBad,
		Error:    errorMessage,
		ReadAt:   readAt,
	}
}

func trimSpace(value string) string {
	return strings.TrimSpace(value)
}
