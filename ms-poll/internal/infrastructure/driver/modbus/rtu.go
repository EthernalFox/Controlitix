package modbus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
	goburrowmodbus "github.com/goburrow/modbus"
)

type RTUDriverFactory struct {
	pool *SerialPortPool
}

func NewRTUDriverFactory(pool *SerialPortPool) *RTUDriverFactory {
	if pool == nil {
		pool = NewSerialPortPool()
	}

	return &RTUDriverFactory{
		pool: pool,
	}
}

func (factory *RTUDriverFactory) SupportedType() string {
	return "modbus_rtu"
}

func (factory *RTUDriverFactory) Build(
	device *domain.DeviceSnapshot,
) (domain.Driver, error) {
	if device == nil {
		return nil, errors.New("device snapshot is nil")
	}

	settings, parseError := parseRTUSettings(device.Settings)
	if parseError != nil {
		return nil, fmt.Errorf("parse rtu settings: %w", parseError)
	}

	return &rtuDriver{
		settings: settings,
		pool:     factory.pool,
	}, nil
}

type rtuSettings struct {
	Port      string `json:"port"`
	BaudRate  int    `json:"baud_rate"`
	DataBits  int    `json:"data_bits"`
	Parity    string `json:"parity"`
	StopBits  int    `json:"stop_bits"`
	UnitID    byte   `json:"unit_id"`
	TimeoutMs int    `json:"timeout_ms"`
}

type rtuDriver struct {
	settings rtuSettings
	pool     *SerialPortPool

	mutex   sync.Mutex
	handler *goburrowmodbus.RTUClientHandler
	client  modbusClient
}

func (driver *rtuDriver) Name() string {
	return "modbus_rtu"
}

func (driver *rtuDriver) Read(
	ctx context.Context,
	tags []*domain.TagSnapshot,
) []domain.Reading {
	lock := driver.pool.Acquire(driver.settings.Port)
	lock.Lock()
	defer lock.Unlock()

	client, prepareError := driver.prepareClient()
	if prepareError != nil {
		return badReadings(tags, prepareError)
	}

	return readWithClient(ctx, tags, client)
}

func (driver *rtuDriver) HealthCheck(ctx context.Context) error {
	if ctxError := ctx.Err(); ctxError != nil {
		return ctxError
	}

	lock := driver.pool.Acquire(driver.settings.Port)
	lock.Lock()
	defer lock.Unlock()

	client, prepareError := driver.prepareClient()
	if prepareError != nil {
		return prepareError
	}

	_, holdingError := client.ReadHoldingRegisters(0, 1)
	if holdingError == nil {
		return nil
	}

	_, coilsError := client.ReadCoils(0, 1)
	if coilsError == nil {
		return nil
	}

	return fmt.Errorf(
		"rtu health check failed: %w",
		errors.Join(holdingError, coilsError),
	)
}

func (driver *rtuDriver) Close() error {
	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	if driver.handler == nil {
		return nil
	}

	closeError := driver.handler.Close()
	driver.handler = nil
	driver.client = nil
	if closeError != nil {
		return fmt.Errorf("close rtu handler: %w", closeError)
	}

	return nil
}

func (driver *rtuDriver) prepareClient() (modbusClient, error) {
	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	if driver.handler == nil {
		handler := goburrowmodbus.NewRTUClientHandler(driver.settings.Port)
		handler.BaudRate = driver.settings.BaudRate
		handler.DataBits = driver.settings.DataBits
		handler.Parity = driver.settings.Parity
		handler.StopBits = driver.settings.StopBits
		handler.Timeout = time.Duration(driver.settings.TimeoutMs) * time.Millisecond
		handler.SlaveId = driver.settings.UnitID

		connectError := handler.Connect()
		if connectError != nil {
			return nil, fmt.Errorf("connect rtu handler: %w", connectError)
		}

		driver.handler = handler
		driver.client = goburrowmodbus.NewClient(handler)
	}

	driver.handler.SlaveId = driver.settings.UnitID
	return driver.client, nil
}

func parseRTUSettings(rawSettings json.RawMessage) (rtuSettings, error) {
	settings := rtuSettings{
		BaudRate:  defaultRTUBaudRate,
		DataBits:  defaultRTUDataBits,
		Parity:    defaultRTUParity,
		StopBits:  defaultRTUStopBits,
		UnitID:    defaultUnitID,
		TimeoutMs: defaultTimeoutMs,
	}

	if len(rawSettings) > 0 {
		unmarshalError := json.Unmarshal(rawSettings, &settings)
		if unmarshalError != nil {
			return rtuSettings{}, fmt.Errorf(
				"unmarshal rtu settings: %w",
				unmarshalError,
			)
		}
	}

	settings.Port = trimSpace(settings.Port)
	if settings.Port == "" {
		return rtuSettings{}, errors.New("port is required for modbus_rtu")
	}
	if settings.BaudRate <= 0 {
		settings.BaudRate = defaultRTUBaudRate
	}
	if settings.DataBits <= 0 {
		settings.DataBits = defaultRTUDataBits
	}
	if settings.StopBits <= 0 {
		settings.StopBits = defaultRTUStopBits
	}
	if settings.TimeoutMs <= 0 {
		settings.TimeoutMs = defaultTimeoutMs
	}
	if settings.UnitID == 0 {
		settings.UnitID = defaultUnitID
	}

	settings.Parity = strings.ToUpper(trimSpace(settings.Parity))
	switch settings.Parity {
	case "N", "E", "O":
	default:
		return rtuSettings{}, fmt.Errorf(
			"unsupported parity value: %s",
			settings.Parity,
		)
	}

	return settings, nil
}

func badReadings(tags []*domain.TagSnapshot, readError error) []domain.Reading {
	readings := make([]domain.Reading, 0, len(tags))
	readAt := time.Now().UTC()
	for _, tag := range tags {
		readings = append(readings, newBadReading(tag, readError, readAt))
	}
	return readings
}
