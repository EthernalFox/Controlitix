package usecase

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

const (
	defaultHealthCheckInterval = 10 * time.Second
	defaultDriverRetryInterval = 30 * time.Second
)

type TagValuesPublisher interface {
	Publish(ctx context.Context, readings []domain.Reading) error
}

type Scheduler struct {
	store     *SnapshotStore
	registry  *DriverRegistry
	publisher TagValuesPublisher
	logger    *slog.Logger

	scanRateResolver   func(tag *domain.TagSnapshot) time.Duration
	healthCheckInterval time.Duration
	driverRetryInterval time.Duration

	mutex   sync.Mutex
	workers map[string]*deviceWorker
}

type deviceWorker struct {
	cancel context.CancelFunc
	done   chan struct{}
}

type namedDriver interface {
	Name() string
}

func NewScheduler(
	store *SnapshotStore,
	registry *DriverRegistry,
	publisher TagValuesPublisher,
	logger *slog.Logger,
) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}

	return &Scheduler{
		store:               store,
		registry:            registry,
		publisher:           publisher,
		logger:              logger,
		scanRateResolver:    ScanRateFor,
		healthCheckInterval: defaultHealthCheckInterval,
		driverRetryInterval: defaultDriverRetryInterval,
		workers:             make(map[string]*deviceWorker),
	}
}

func (scheduler *Scheduler) Run(ctx context.Context) error {
	if ctx == nil {
		return errors.New("scheduler context is nil")
	}
	if scheduler.store == nil {
		return errors.New("snapshot store is not configured")
	}
	if scheduler.registry == nil {
		return errors.New("driver registry is not configured")
	}
	if scheduler.publisher == nil {
		return errors.New("tag values publisher is not configured")
	}

	events := scheduler.store.Subscribe()
	defer scheduler.store.Unsubscribe(events)

	scheduler.logger.Info("scheduler started")
	defer func() {
		scheduler.stopAllDevices()
		scheduler.logger.Info("scheduler stopped")
	}()

	scheduler.reinitializeAllDevices(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, isOpen := <-events:
			if !isOpen {
				return nil
			}
			scheduler.handleSnapshotEvent(ctx, event)
		}
	}
}

func (scheduler *Scheduler) handleSnapshotEvent(
	ctx context.Context,
	event SnapshotEvent,
) {
	switch event.Kind {
	case "replaced":
		scheduler.reinitializeAllDevices(ctx)
	case "device_upserted":
		scheduler.restartDevice(ctx, event.DeviceID)
	case "device_removed":
		scheduler.stopDevice(event.DeviceID)
	case "tag_upserted", "tag_removed":
		deviceID := event.DeviceID
		if deviceID == "" {
			scheduler.reinitializeAllDevices(ctx)
			return
		}
		scheduler.restartDevice(ctx, deviceID)
	default:
		scheduler.logger.Debug("unknown snapshot event", "kind", event.Kind)
	}
}

func (scheduler *Scheduler) reinitializeAllDevices(ctx context.Context) {
	desiredDeviceIDs := scheduler.desiredDeviceIDs()

	for _, runningDeviceID := range scheduler.runningDeviceIDs() {
		scheduler.stopDevice(runningDeviceID)
	}

	sort.Strings(desiredDeviceIDs)
	for _, deviceID := range desiredDeviceIDs {
		scheduler.startDevice(ctx, deviceID)
	}
}

func (scheduler *Scheduler) restartDevice(ctx context.Context, deviceID string) {
	if deviceID == "" {
		return
	}

	scheduler.stopDevice(deviceID)
	scheduler.startDevice(ctx, deviceID)
}

func (scheduler *Scheduler) startDevice(ctx context.Context, deviceID string) {
	if deviceID == "" {
		return
	}

	_, tags, exists := scheduler.getDeviceConfig(deviceID)
	if !exists || len(tags) == 0 {
		return
	}

	deviceContext, cancelDevice := context.WithCancel(ctx)
	deviceDone := make(chan struct{})

	scheduler.mutex.Lock()
	scheduler.workers[deviceID] = &deviceWorker{
		cancel: cancelDevice,
		done:   deviceDone,
	}
	scheduler.mutex.Unlock()

	go func() {
		defer close(deviceDone)
		scheduler.runDevice(deviceContext, deviceID)
	}()
}

func (scheduler *Scheduler) stopDevice(deviceID string) {
	if deviceID == "" {
		return
	}

	scheduler.mutex.Lock()
	worker, exists := scheduler.workers[deviceID]
	if exists {
		delete(scheduler.workers, deviceID)
	}
	scheduler.mutex.Unlock()

	if !exists {
		return
	}

	worker.cancel()
	<-worker.done
}

func (scheduler *Scheduler) stopAllDevices() {
	runningDeviceIDs := scheduler.runningDeviceIDs()
	for _, deviceID := range runningDeviceIDs {
		scheduler.stopDevice(deviceID)
	}
}

func (scheduler *Scheduler) runningDeviceIDs() []string {
	scheduler.mutex.Lock()
	defer scheduler.mutex.Unlock()

	deviceIDs := make([]string, 0, len(scheduler.workers))
	for deviceID := range scheduler.workers {
		deviceIDs = append(deviceIDs, deviceID)
	}

	return deviceIDs
}

func (scheduler *Scheduler) desiredDeviceIDs() []string {
	snapshot := scheduler.store.Get()
	if snapshot == nil {
		return nil
	}

	deviceIDs := make([]string, 0, len(snapshot.Devices))
	for deviceID := range snapshot.Devices {
		if len(snapshot.TagsByDevice[deviceID]) == 0 {
			continue
		}

		deviceIDs = append(deviceIDs, deviceID)
	}

	return deviceIDs
}

func (scheduler *Scheduler) runDevice(ctx context.Context, deviceID string) {
	for {
		deviceSnapshot, tagSnapshots, exists := scheduler.getDeviceConfig(deviceID)
		if !exists || len(tagSnapshots) == 0 {
			return
		}

		driver, buildError := scheduler.registry.Build(deviceSnapshot)
		if buildError != nil {
			scheduler.logger.Warn(
				"device driver build failed",
				"device_id",
				deviceID,
				"device_type",
				deviceSnapshot.TypeName,
				"error",
				buildError,
			)

			select {
			case <-ctx.Done():
				return
			case <-time.After(scheduler.driverRetryInterval):
				continue
			}
		}

		driverName := "unknown"
		if named, isNamed := driver.(namedDriver); isNamed {
			driverName = named.Name()
		}

		scheduler.logger.Info(
			"device started",
			"device_id",
			deviceID,
			"tags",
			len(tagSnapshots),
			"driver",
			driverName,
		)

		scheduler.runDeviceLoop(ctx, deviceSnapshot, tagSnapshots, driver)

		if closeError := driver.Close(); closeError != nil {
			scheduler.logger.Error(
				"failed to close device driver",
				"device_id",
				deviceID,
				"error",
				closeError,
			)
		}

		scheduler.logger.Info("device stopped", "device_id", deviceID)
		return
	}
}

func (scheduler *Scheduler) runDeviceLoop(
	ctx context.Context,
	deviceSnapshot *domain.DeviceSnapshot,
	tagSnapshots []*domain.TagSnapshot,
	driver domain.Driver,
) {
	tagGroups := scheduler.groupTagsByScanRate(tagSnapshots)
	if len(tagGroups) == 0 {
		return
	}

	healthTicker := time.NewTicker(scheduler.healthCheckInterval)
	defer healthTicker.Stop()

	type rateCase struct {
		tags   []*domain.TagSnapshot
		ticker *time.Ticker
	}

	rateCases := make([]rateCase, 0, len(tagGroups))
	selectCases := []reflect.SelectCase{
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		},
		{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(healthTicker.C),
		},
	}

	for scanRate, groupedTags := range tagGroups {
		ticker := time.NewTicker(scanRate)
		rateCases = append(rateCases, rateCase{
			tags:   groupedTags,
			ticker: ticker,
		})
		selectCases = append(selectCases, reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ticker.C),
		})
	}

	defer func() {
		for _, rateCase := range rateCases {
			rateCase.ticker.Stop()
		}
	}()

	deviceHealthy := true

	for {
		selectedCase, _, isOpen := reflect.Select(selectCases)
		if !isOpen {
			return
		}

		switch selectedCase {
		case 0:
			return
		case 1:
			healthError := driver.HealthCheck(ctx)
			if healthError != nil {
				deviceHealthy = false
				scheduler.logger.Warn(
					"device health check failed",
					"device_id",
					deviceSnapshot.ID,
					"error",
					healthError,
				)
				continue
			}

			deviceHealthy = true
		default:
			if !deviceHealthy {
				continue
			}

			tags := rateCases[selectedCase-2].tags
			readings := driver.Read(ctx, tags)
			if len(readings) == 0 {
				continue
			}

			publishError := scheduler.publisher.Publish(ctx, readings)
			if publishError != nil {
				scheduler.logger.Error(
					"failed to publish readings",
					"device_id",
					deviceSnapshot.ID,
					"error",
					publishError,
				)
			}
		}
	}
}

func (scheduler *Scheduler) getDeviceConfig(
	deviceID string,
) (*domain.DeviceSnapshot, []*domain.TagSnapshot, bool) {
	snapshot := scheduler.store.Get()
	if snapshot == nil {
		return nil, nil, false
	}

	deviceSnapshot, exists := snapshot.Devices[deviceID]
	if !exists {
		return nil, nil, false
	}

	tagSnapshots := snapshot.TagsByDevice[deviceID]
	if len(tagSnapshots) == 0 {
		return cloneDeviceSnapshot(deviceSnapshot), nil, true
	}

	clonedTagSnapshots := make([]*domain.TagSnapshot, 0, len(tagSnapshots))
	for _, tagSnapshot := range tagSnapshots {
		if tagSnapshot == nil {
			continue
		}

		clonedTagSnapshots = append(clonedTagSnapshots, cloneTagSnapshot(tagSnapshot))
	}

	return cloneDeviceSnapshot(deviceSnapshot), clonedTagSnapshots, true
}

func (scheduler *Scheduler) groupTagsByScanRate(
	tagSnapshots []*domain.TagSnapshot,
) map[time.Duration][]*domain.TagSnapshot {
	groups := make(map[time.Duration][]*domain.TagSnapshot)
	for _, tagSnapshot := range tagSnapshots {
		if tagSnapshot == nil {
			continue
		}

		scanRate := scheduler.scanRateResolver(tagSnapshot)
		if scanRate <= 0 {
			continue
		}

		groups[scanRate] = append(groups[scanRate], tagSnapshot)
	}

	return groups
}
