package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/infrastructure/repository"
)

type ConfigLoader struct {
	repository configLoaderRepository
	store      *SnapshotStore
	logger     *slog.Logger
}

func NewConfigLoader(
	repository *repository.ConfigRepository,
	store *SnapshotStore,
	logger *slog.Logger,
) *ConfigLoader {
	return newConfigLoader(repository, store, logger)
}

func newConfigLoader(
	repository configLoaderRepository,
	store *SnapshotStore,
	logger *slog.Logger,
) *ConfigLoader {
	if logger == nil {
		logger = slog.Default()
	}

	return &ConfigLoader{
		repository: repository,
		store:      store,
		logger:     logger,
	}
}

func (loader *ConfigLoader) Load(ctx context.Context) error {
	startedAt := time.Now()

	snapshot, loadError := loader.repository.LoadAll(ctx)
	if loadError != nil {
		return fmt.Errorf("load full snapshot: %w", loadError)
	}
	if snapshot == nil {
		snapshot = ensureSnapshot(nil)
	}

	snapshot.LoadedAt = time.Now().UTC()
	loader.store.Replace(snapshot)

	loader.logger.Info(
		"config snapshot loaded",
		"devices_count",
		len(snapshot.Devices),
		"tags_count",
		len(snapshot.Tags),
		"duration",
		time.Since(startedAt),
	)

	return nil
}
