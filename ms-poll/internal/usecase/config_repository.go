package usecase

import (
	"context"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

type configLoaderRepository interface {
	LoadAll(ctx context.Context) (*domain.Snapshot, error)
}

type configEventRepository interface {
	LoadDevice(ctx context.Context, id string) (*domain.DeviceSnapshot, error)
	LoadTag(ctx context.Context, id string) (*domain.TagSnapshot, error)
}
