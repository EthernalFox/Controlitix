package usecase

import (
	"context"
	"sync"
	"time"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

type fakeDriverFactory struct {
	supportedType string

	mutex       sync.Mutex
	buildCount  int
	builtDriver []*fakeDriver

	createDriver func(device *domain.DeviceSnapshot) *fakeDriver
}

func newFakeDriverFactory(supportedType string) *fakeDriverFactory {
	return &fakeDriverFactory{
		supportedType: supportedType,
		createDriver: func(device *domain.DeviceSnapshot) *fakeDriver {
			return newFakeDriver(device.ID)
		},
	}
}

func (factory *fakeDriverFactory) Build(
	device *domain.DeviceSnapshot,
) (domain.Driver, error) {
	driver := factory.createDriver(device)

	factory.mutex.Lock()
	factory.buildCount++
	factory.builtDriver = append(factory.builtDriver, driver)
	factory.mutex.Unlock()

	return driver, nil
}

func (factory *fakeDriverFactory) SupportedType() string {
	return factory.supportedType
}

func (factory *fakeDriverFactory) BuildsCount() int {
	factory.mutex.Lock()
	defer factory.mutex.Unlock()

	return factory.buildCount
}

func (factory *fakeDriverFactory) Drivers() []*fakeDriver {
	factory.mutex.Lock()
	defer factory.mutex.Unlock()

	result := make([]*fakeDriver, 0, len(factory.builtDriver))
	result = append(result, factory.builtDriver...)
	return result
}

type fakeDriver struct {
	deviceID string

	mutex          sync.Mutex
	readCalls      int
	healthCalls    int
	closeCalls     int
	readTagCounts  []int
	readTagHistory [][]string
	healthSequence []error
	readSignal     chan struct{}
	closedSignal   chan struct{}
}

func newFakeDriver(deviceID string) *fakeDriver {
	return &fakeDriver{
		deviceID:     deviceID,
		readSignal:   make(chan struct{}, 128),
		closedSignal: make(chan struct{}, 16),
	}
}

func (driver *fakeDriver) Name() string {
	return "noop"
}

func (driver *fakeDriver) Read(
	_ context.Context,
	tags []*domain.TagSnapshot,
) []domain.Reading {
	tagIDs := make([]string, 0, len(tags))
	driver.mutex.Lock()
	driver.readCalls++
	driver.readTagCounts = append(driver.readTagCounts, len(tags))
	for _, tag := range tags {
		if tag == nil {
			continue
		}
		tagIDs = append(tagIDs, tag.ID)
	}
	driver.readTagHistory = append(driver.readTagHistory, append([]string(nil), tagIDs...))
	driver.mutex.Unlock()

	select {
	case driver.readSignal <- struct{}{}:
	default:
	}

	readings := make([]domain.Reading, 0, len(tags))
	for _, tag := range tags {
		if tag == nil {
			continue
		}

		readings = append(readings, domain.Reading{
			TagID:    tag.ID,
			DeviceID: driver.deviceID,
			Quality:  domain.QualityGood,
			ReadAt:   time.Now().UTC(),
		})
	}

	return readings
}

func (driver *fakeDriver) HealthCheck(_ context.Context) error {
	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	driver.healthCalls++
	if len(driver.healthSequence) == 0 {
		return nil
	}

	result := driver.healthSequence[0]
	driver.healthSequence = driver.healthSequence[1:]
	return result
}

func (driver *fakeDriver) Close() error {
	driver.mutex.Lock()
	driver.closeCalls++
	driver.mutex.Unlock()

	select {
	case driver.closedSignal <- struct{}{}:
	default:
	}

	return nil
}

func (driver *fakeDriver) SetHealthSequence(sequence ...error) {
	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	driver.healthSequence = append([]error(nil), sequence...)
}

func (driver *fakeDriver) ReadCalls() int {
	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	return driver.readCalls
}

func (driver *fakeDriver) HealthCalls() int {
	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	return driver.healthCalls
}

func (driver *fakeDriver) CloseCalls() int {
	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	return driver.closeCalls
}

func (driver *fakeDriver) ReadTagHistory() [][]string {
	driver.mutex.Lock()
	defer driver.mutex.Unlock()

	history := make([][]string, 0, len(driver.readTagHistory))
	for _, entry := range driver.readTagHistory {
		history = append(history, append([]string(nil), entry...))
	}
	return history
}

type fakePublisher struct {
	mutex sync.Mutex

	publishCalls int
	published    [][]domain.Reading
}

func (publisher *fakePublisher) Publish(
	_ context.Context,
	readings []domain.Reading,
) error {
	publisher.mutex.Lock()
	defer publisher.mutex.Unlock()

	publisher.publishCalls++
	copiedReadings := append([]domain.Reading(nil), readings...)
	publisher.published = append(publisher.published, copiedReadings)

	return nil
}

func (publisher *fakePublisher) PublishCalls() int {
	publisher.mutex.Lock()
	defer publisher.mutex.Unlock()

	return publisher.publishCalls
}
