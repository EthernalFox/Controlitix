package domain

import "context"

type Driver interface {
	Read(ctx context.Context, tags []*TagSnapshot) []Reading
	HealthCheck(ctx context.Context) error
	Close() error
}

type DriverFactory interface {
	Build(device *DeviceSnapshot) (Driver, error)
	SupportedType() string
}
