package usecase

import (
	"fmt"
	"sync"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

type DriverRegistry struct {
	mutex     sync.RWMutex
	factories map[string]domain.DriverFactory
}

func NewDriverRegistry() *DriverRegistry {
	return &DriverRegistry{
		factories: make(map[string]domain.DriverFactory),
	}
}

func (registry *DriverRegistry) Register(factory domain.DriverFactory) {
	if factory == nil {
		return
	}

	supportedType := factory.SupportedType()
	if supportedType == "" {
		return
	}

	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	registry.factories[supportedType] = factory
}

func (registry *DriverRegistry) Build(
	device *domain.DeviceSnapshot,
) (domain.Driver, error) {
	if device == nil {
		return nil, fmt.Errorf("build driver: device is nil")
	}

	registry.mutex.RLock()
	factory, exists := registry.factories[device.TypeName]
	registry.mutex.RUnlock()
	if !exists {
		return nil, fmt.Errorf("unsupported device type: %s", device.TypeName)
	}

	driver, buildError := factory.Build(device)
	if buildError != nil {
		return nil, fmt.Errorf(
			"build driver for device type %s: %w",
			device.TypeName,
			buildError,
		)
	}

	return driver, nil
}
