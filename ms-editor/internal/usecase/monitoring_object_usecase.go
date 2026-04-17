package usecase

import (
	"context"
	"strings"

	"github.com/EthernalFox/Controlitix/ms-editor/internal/domain"
)

type MonitoringObjectUseCase struct {
	monitoringObjectRepository domain.MonitoringObjectRepository
}

func NewMonitoringObjectUseCase(
	monitoringObjectRepository domain.MonitoringObjectRepository,
) *MonitoringObjectUseCase {
	return &MonitoringObjectUseCase{
		monitoringObjectRepository: monitoringObjectRepository,
	}
}

func (useCase *MonitoringObjectUseCase) CreateMonitoringObject(
	ctx context.Context,
	name string,
	description *string,
) (domain.MonitoringObject, error) {
	if strings.TrimSpace(name) == "" {
		return domain.MonitoringObject{}, domain.ErrInvalidInput
	}

	monitoringObject := domain.MonitoringObject{
		Name:        name,
		Description: description,
	}

	return useCase.monitoringObjectRepository.CreateMonitoringObject(
		ctx,
		monitoringObject,
	)
}

func (useCase *MonitoringObjectUseCase) GetMonitoringObject(
	ctx context.Context,
	monitoringObjectID string,
) (domain.MonitoringObject, error) {
	if strings.TrimSpace(monitoringObjectID) == "" {
		return domain.MonitoringObject{}, domain.ErrInvalidInput
	}

	return useCase.monitoringObjectRepository.GetMonitoringObject(
		ctx,
		monitoringObjectID,
	)
}

func (useCase *MonitoringObjectUseCase) UpdateMonitoringObject(
	ctx context.Context,
	monitoringObjectID string,
	name *string,
	description *string,
) (domain.MonitoringObject, error) {
	if strings.TrimSpace(monitoringObjectID) == "" {
		return domain.MonitoringObject{}, domain.ErrInvalidInput
	}

	update := domain.MonitoringObjectUpdate{
		Name:        name,
		Description: description,
	}

	return useCase.monitoringObjectRepository.UpdateMonitoringObject(
		ctx,
		monitoringObjectID,
		update,
	)
}

func (useCase *MonitoringObjectUseCase) DeleteMonitoringObject(
	ctx context.Context,
	monitoringObjectID string,
) error {
	if strings.TrimSpace(monitoringObjectID) == "" {
		return domain.ErrInvalidInput
	}

	return useCase.monitoringObjectRepository.DeleteMonitoringObject(
		ctx,
		monitoringObjectID,
	)
}

func (useCase *MonitoringObjectUseCase) ListMonitoringObjects(
	ctx context.Context,
	query domain.ObjectListQuery,
) (domain.ListResult[domain.MonitoringObject], error) {
	return useCase.monitoringObjectRepository.ListMonitoringObjects(ctx, query)
}
