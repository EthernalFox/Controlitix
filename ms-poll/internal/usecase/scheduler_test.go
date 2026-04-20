package usecase

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

func TestSchedulerDeviceWithoutTagsDoesNotStartWorker(t *testing.T) {
	store := NewSnapshotStore()
	store.Replace(buildSnapshot(
		[]*domain.DeviceSnapshot{
			{
				ID:        "device-1",
				TypeName:  "modbus_tcp",
				TypeID:    1,
				Name:      "Device 1",
				UpdatedAt: time.Now().UTC(),
			},
		},
		nil,
	))

	factory := newFakeDriverFactory("modbus_tcp")
	registry := NewDriverRegistry()
	registry.Register(factory)

	scheduler := NewScheduler(store, registry, &fakePublisher{}, nil)
	cancelScheduler, done := runScheduler(t, scheduler)
	defer stopScheduler(t, cancelScheduler, done)

	time.Sleep(80 * time.Millisecond)

	if factory.BuildsCount() != 0 {
		t.Fatalf("expected zero driver builds, got %d", factory.BuildsCount())
	}
}

func TestSchedulerRunsDifferentScanRatesForOneDevice(t *testing.T) {
	store := NewSnapshotStore()
	store.Replace(buildSnapshot(
		[]*domain.DeviceSnapshot{
			{
				ID:        "device-1",
				TypeName:  "modbus_tcp",
				TypeID:    1,
				Name:      "Device 1",
				UpdatedAt: time.Now().UTC(),
			},
		},
		[]*domain.TagSnapshot{
			{
				ID:         "tag-fast",
				DeviceID:   "device-1",
				Name:       "Fast",
				DataType:   "float64",
				UnitSymbol: pointerToString("V"),
				UpdatedAt:  time.Now().UTC(),
			},
			{
				ID:         "tag-slow",
				DeviceID:   "device-1",
				Name:       "Slow",
				DataType:   "float64",
				UnitSymbol: pointerToString("°C"),
				UpdatedAt:  time.Now().UTC(),
			},
		},
	))

	factory := newFakeDriverFactory("modbus_tcp")
	registry := NewDriverRegistry()
	registry.Register(factory)

	publisher := &fakePublisher{}
	scheduler := NewScheduler(store, registry, publisher, nil)
	scheduler.scanRateResolver = func(tag *domain.TagSnapshot) time.Duration {
		if tag != nil && tag.UnitSymbol != nil && *tag.UnitSymbol == "°C" {
			return 70 * time.Millisecond
		}
		return 20 * time.Millisecond
	}
	scheduler.healthCheckInterval = time.Hour

	cancelScheduler, done := runScheduler(t, scheduler)
	defer stopScheduler(t, cancelScheduler, done)

	if !waitForCondition(500*time.Millisecond, 10*time.Millisecond, func() bool {
		return factory.BuildsCount() == 1
	}) {
		t.Fatal("driver was not built")
	}

	time.Sleep(260 * time.Millisecond)

	drivers := factory.Drivers()
	if len(drivers) != 1 {
		t.Fatalf("expected one driver, got %d", len(drivers))
	}

	readHistory := drivers[0].ReadTagHistory()
	fastReads := 0
	slowReads := 0
	for _, readTags := range readHistory {
		if len(readTags) != 1 {
			continue
		}

		switch readTags[0] {
		case "tag-fast":
			fastReads++
		case "tag-slow":
			slowReads++
		}
	}

	if fastReads < 4 {
		t.Fatalf("expected at least 4 fast reads, got %d", fastReads)
	}
	if slowReads < 2 {
		t.Fatalf("expected at least 2 slow reads, got %d", slowReads)
	}
	if fastReads <= slowReads {
		t.Fatalf("expected fast reads > slow reads, got %d and %d", fastReads, slowReads)
	}
}

func TestSchedulerStopsDeviceOnDeviceRemoved(t *testing.T) {
	store := NewSnapshotStore()
	store.Replace(buildSnapshot(
		[]*domain.DeviceSnapshot{
			{
				ID:        "device-1",
				TypeName:  "modbus_tcp",
				TypeID:    1,
				Name:      "Device 1",
				UpdatedAt: time.Now().UTC(),
			},
		},
		[]*domain.TagSnapshot{
			{
				ID:        "tag-1",
				DeviceID:  "device-1",
				Name:      "Tag 1",
				DataType:  "float64",
				UpdatedAt: time.Now().UTC(),
			},
		},
	))

	factory := newFakeDriverFactory("modbus_tcp")
	registry := NewDriverRegistry()
	registry.Register(factory)

	scheduler := NewScheduler(store, registry, &fakePublisher{}, nil)
	scheduler.scanRateResolver = func(_ *domain.TagSnapshot) time.Duration {
		return 20 * time.Millisecond
	}
	scheduler.healthCheckInterval = time.Hour

	cancelScheduler, done := runScheduler(t, scheduler)
	defer stopScheduler(t, cancelScheduler, done)

	if !waitForCondition(500*time.Millisecond, 10*time.Millisecond, func() bool {
		return factory.BuildsCount() == 1
	}) {
		t.Fatal("driver was not built")
	}

	driver := factory.Drivers()[0]
	if !waitForCondition(500*time.Millisecond, 10*time.Millisecond, func() bool {
		return driver.ReadCalls() > 0
	}) {
		t.Fatal("driver did not receive read calls")
	}

	store.RemoveDevice("device-1")

	if !waitForCondition(500*time.Millisecond, 10*time.Millisecond, func() bool {
		return driver.CloseCalls() == 1
	}) {
		t.Fatal("driver close was not called once after device removal")
	}
}

func TestSchedulerRestartsDriverOnDeviceUpsert(t *testing.T) {
	store := NewSnapshotStore()
	store.Replace(buildSnapshot(
		[]*domain.DeviceSnapshot{
			{
				ID:        "device-1",
				TypeName:  "modbus_tcp",
				TypeID:    1,
				Name:      "Device 1",
				UpdatedAt: time.Now().UTC(),
			},
		},
		[]*domain.TagSnapshot{
			{
				ID:        "tag-1",
				DeviceID:  "device-1",
				Name:      "Tag 1",
				DataType:  "float64",
				UpdatedAt: time.Now().UTC(),
			},
		},
	))

	factory := newFakeDriverFactory("modbus_tcp")
	registry := NewDriverRegistry()
	registry.Register(factory)

	scheduler := NewScheduler(store, registry, &fakePublisher{}, nil)
	scheduler.scanRateResolver = func(_ *domain.TagSnapshot) time.Duration {
		return 20 * time.Millisecond
	}
	scheduler.healthCheckInterval = time.Hour

	cancelScheduler, done := runScheduler(t, scheduler)
	defer stopScheduler(t, cancelScheduler, done)

	if !waitForCondition(500*time.Millisecond, 10*time.Millisecond, func() bool {
		return factory.BuildsCount() == 1
	}) {
		t.Fatal("driver was not built")
	}

	firstDriver := factory.Drivers()[0]
	store.UpsertDevice(&domain.DeviceSnapshot{
		ID:        "device-1",
		TypeName:  "modbus_tcp",
		TypeID:    1,
		Name:      "Device 1 updated",
		UpdatedAt: time.Now().UTC(),
	})

	if !waitForCondition(500*time.Millisecond, 10*time.Millisecond, func() bool {
		return factory.BuildsCount() == 2
	}) {
		t.Fatal("expected driver rebuild after device upsert")
	}

	if !waitForCondition(500*time.Millisecond, 10*time.Millisecond, func() bool {
		return firstDriver.CloseCalls() == 1
	}) {
		t.Fatal("expected first driver to be closed once")
	}
}

func TestSchedulerRestartsDeviceOnTagRemoved(t *testing.T) {
	store := NewSnapshotStore()
	store.Replace(buildSnapshot(
		[]*domain.DeviceSnapshot{
			{
				ID:        "device-1",
				TypeName:  "modbus_tcp",
				TypeID:    1,
				Name:      "Device 1",
				UpdatedAt: time.Now().UTC(),
			},
		},
		[]*domain.TagSnapshot{
			{
				ID:        "tag-1",
				DeviceID:  "device-1",
				Name:      "Tag 1",
				DataType:  "float64",
				UpdatedAt: time.Now().UTC(),
			},
			{
				ID:        "tag-2",
				DeviceID:  "device-1",
				Name:      "Tag 2",
				DataType:  "float64",
				UpdatedAt: time.Now().UTC(),
			},
		},
	))

	factory := newFakeDriverFactory("modbus_tcp")
	registry := NewDriverRegistry()
	registry.Register(factory)

	scheduler := NewScheduler(store, registry, &fakePublisher{}, nil)
	scheduler.scanRateResolver = func(_ *domain.TagSnapshot) time.Duration {
		return 20 * time.Millisecond
	}
	scheduler.healthCheckInterval = time.Hour

	cancelScheduler, done := runScheduler(t, scheduler)
	defer stopScheduler(t, cancelScheduler, done)

	if !waitForCondition(500*time.Millisecond, 10*time.Millisecond, func() bool {
		return factory.BuildsCount() == 1
	}) {
		t.Fatal("driver was not built")
	}

	firstDriver := factory.Drivers()[0]
	store.RemoveTag("tag-2")

	if !waitForCondition(500*time.Millisecond, 10*time.Millisecond, func() bool {
		return factory.BuildsCount() == 2
	}) {
		t.Fatal("expected driver rebuild after tag removal")
	}
	if !waitForCondition(500*time.Millisecond, 10*time.Millisecond, func() bool {
		return firstDriver.CloseCalls() == 1
	}) {
		t.Fatal("expected first driver to be closed")
	}

	secondDriver := factory.Drivers()[1]
	if !waitForCondition(500*time.Millisecond, 10*time.Millisecond, func() bool {
		return secondDriver.ReadCalls() > 0
	}) {
		t.Fatal("expected second driver to perform reads")
	}

	for _, readTags := range secondDriver.ReadTagHistory() {
		if len(readTags) != 1 || readTags[0] != "tag-1" {
			t.Fatalf("expected restarted driver to read only tag-1, got %v", readTags)
		}
	}
}

func TestSchedulerSkipsReadAfterHealthErrorAndRecovers(t *testing.T) {
	store := NewSnapshotStore()
	store.Replace(buildSnapshot(
		[]*domain.DeviceSnapshot{
			{
				ID:        "device-1",
				TypeName:  "modbus_tcp",
				TypeID:    1,
				Name:      "Device 1",
				UpdatedAt: time.Now().UTC(),
			},
		},
		[]*domain.TagSnapshot{
			{
				ID:        "tag-1",
				DeviceID:  "device-1",
				Name:      "Tag 1",
				DataType:  "float64",
				UpdatedAt: time.Now().UTC(),
			},
		},
	))

	factory := newFakeDriverFactory("modbus_tcp")
	factory.createDriver = func(device *domain.DeviceSnapshot) *fakeDriver {
		driver := newFakeDriver(device.ID)
		driver.SetHealthSequence(errors.New("health failed"), nil)
		return driver
	}

	registry := NewDriverRegistry()
	registry.Register(factory)

	scheduler := NewScheduler(store, registry, &fakePublisher{}, nil)
	scheduler.scanRateResolver = func(_ *domain.TagSnapshot) time.Duration {
		return 10 * time.Millisecond
	}
	scheduler.healthCheckInterval = 30 * time.Millisecond

	cancelScheduler, done := runScheduler(t, scheduler)
	defer stopScheduler(t, cancelScheduler, done)

	if !waitForCondition(500*time.Millisecond, 10*time.Millisecond, func() bool {
		return factory.BuildsCount() == 1
	}) {
		t.Fatal("driver was not built")
	}

	driver := factory.Drivers()[0]

	if !waitForCondition(500*time.Millisecond, 5*time.Millisecond, func() bool {
		return driver.HealthCalls() >= 1
	}) {
		t.Fatal("expected first health check call")
	}

	time.Sleep(5 * time.Millisecond)
	readCountAfterFailure := driver.ReadCalls()
	time.Sleep(15 * time.Millisecond)
	if driver.ReadCalls() != readCountAfterFailure {
		t.Fatal("expected reads to be skipped while device is unhealthy")
	}

	if !waitForCondition(500*time.Millisecond, 5*time.Millisecond, func() bool {
		return driver.HealthCalls() >= 2
	}) {
		t.Fatal("expected recovery health check call")
	}

	if !waitForCondition(500*time.Millisecond, 5*time.Millisecond, func() bool {
		return driver.ReadCalls() > readCountAfterFailure
	}) {
		t.Fatal("expected reads to resume after health recovery")
	}
}

func TestSchedulerClosesDriversOnContextDone(t *testing.T) {
	store := NewSnapshotStore()
	store.Replace(buildSnapshot(
		[]*domain.DeviceSnapshot{
			{
				ID:        "device-1",
				TypeName:  "modbus_tcp",
				TypeID:    1,
				Name:      "Device 1",
				UpdatedAt: time.Now().UTC(),
			},
			{
				ID:        "device-2",
				TypeName:  "modbus_tcp",
				TypeID:    1,
				Name:      "Device 2",
				UpdatedAt: time.Now().UTC(),
			},
		},
		[]*domain.TagSnapshot{
			{
				ID:        "tag-1",
				DeviceID:  "device-1",
				Name:      "Tag 1",
				DataType:  "float64",
				UpdatedAt: time.Now().UTC(),
			},
			{
				ID:        "tag-2",
				DeviceID:  "device-2",
				Name:      "Tag 2",
				DataType:  "float64",
				UpdatedAt: time.Now().UTC(),
			},
		},
	))

	factory := newFakeDriverFactory("modbus_tcp")
	registry := NewDriverRegistry()
	registry.Register(factory)

	scheduler := NewScheduler(store, registry, &fakePublisher{}, nil)
	scheduler.scanRateResolver = func(_ *domain.TagSnapshot) time.Duration {
		return 15 * time.Millisecond
	}
	scheduler.healthCheckInterval = time.Hour

	goroutinesBefore := runtime.NumGoroutine()
	cancelScheduler, done := runScheduler(t, scheduler)

	if !waitForCondition(500*time.Millisecond, 10*time.Millisecond, func() bool {
		return factory.BuildsCount() == 2
	}) {
		t.Fatal("expected two drivers to be built")
	}

	stopScheduler(t, cancelScheduler, done)

	for _, driver := range factory.Drivers() {
		if driver.CloseCalls() != 1 {
			t.Fatalf("expected driver close call count 1, got %d", driver.CloseCalls())
		}
	}

	time.Sleep(80 * time.Millisecond)
	goroutinesAfter := runtime.NumGoroutine()
	if goroutinesAfter > goroutinesBefore+8 {
		t.Fatalf(
			"expected goroutines to stabilize, before=%d after=%d",
			goroutinesBefore,
			goroutinesAfter,
		)
	}
}

func runScheduler(
	t *testing.T,
	scheduler *Scheduler,
) (context.CancelFunc, <-chan error) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- scheduler.Run(ctx)
	}()

	return cancel, done
}

func stopScheduler(
	t *testing.T,
	cancel context.CancelFunc,
	done <-chan error,
) {
	t.Helper()

	cancel()
	select {
	case runError := <-done:
		if runError != nil {
			t.Fatalf("scheduler exited with error: %v", runError)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("scheduler did not stop in time")
	}
}

func waitForCondition(
	timeout time.Duration,
	interval time.Duration,
	check func() bool,
) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return true
		}

		time.Sleep(interval)
	}

	return check()
}

func pointerToString(value string) *string {
	return &value
}
