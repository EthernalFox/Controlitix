package repository

import (
	"context"
	"database/sql"

	"gitlab.controlitix.ru/controlitix/ms-editor/internal/domain"
)

type PostgresRepository struct {
	databaseConnection *sql.DB
}

func NewPostgresRepository(databaseConnection *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		databaseConnection: databaseConnection,
	}
}

func (repository *PostgresRepository) CreateMonitoringObject(
	ctx context.Context,
	monitoringObject domain.MonitoringObject,
) (domain.MonitoringObject, error) {
	return domain.MonitoringObject{}, domain.ErrNotImplemented
}

func (repository *PostgresRepository) GetMonitoringObject(
	ctx context.Context,
	monitoringObjectID string,
) (domain.MonitoringObject, error) {
	return domain.MonitoringObject{}, domain.ErrNotImplemented
}

func (repository *PostgresRepository) UpdateMonitoringObject(
	ctx context.Context,
	monitoringObjectID string,
	update domain.MonitoringObjectUpdate,
) (domain.MonitoringObject, error) {
	return domain.MonitoringObject{}, domain.ErrNotImplemented
}

func (repository *PostgresRepository) DeleteMonitoringObject(
	ctx context.Context,
	monitoringObjectID string,
) error {
	return domain.ErrNotImplemented
}

func (repository *PostgresRepository) ListMonitoringObjects(
	ctx context.Context,
) ([]domain.MonitoringObject, error) {
	return nil, domain.ErrNotImplemented
}

func (repository *PostgresRepository) CreateDiagram(
	ctx context.Context,
	diagram domain.Diagram,
) (domain.Diagram, error) {
	return domain.Diagram{}, domain.ErrNotImplemented
}

func (repository *PostgresRepository) GetDiagram(
	ctx context.Context,
	diagramID string,
) (domain.Diagram, error) {
	return domain.Diagram{}, domain.ErrNotImplemented
}

func (repository *PostgresRepository) UpdateDiagram(
	ctx context.Context,
	diagramID string,
	update domain.DiagramUpdate,
) (domain.Diagram, error) {
	return domain.Diagram{}, domain.ErrNotImplemented
}

func (repository *PostgresRepository) DeleteDiagram(
	ctx context.Context,
	diagramID string,
) error {
	return domain.ErrNotImplemented
}

func (repository *PostgresRepository) PublishDiagram(
	ctx context.Context,
	diagramID string,
) (domain.Diagram, error) {
	return domain.Diagram{}, domain.ErrNotImplemented
}

func (repository *PostgresRepository) ListDiagramsByMonitoringObject(
	ctx context.Context,
	monitoringObjectID string,
) ([]domain.Diagram, error) {
	return nil, domain.ErrNotImplemented
}

func (repository *PostgresRepository) CreateFigures(
	ctx context.Context,
	diagramID string,
	figures []domain.Figure,
) ([]domain.Figure, error) {
	return nil, domain.ErrNotImplemented
}

func (repository *PostgresRepository) UpdateFigure(
	ctx context.Context,
	figureID string,
	update domain.FigureUpdate,
) (domain.Figure, error) {
	return domain.Figure{}, domain.ErrNotImplemented
}

func (repository *PostgresRepository) DeleteFigure(
	ctx context.Context,
	figureID string,
) error {
	return domain.ErrNotImplemented
}

func (repository *PostgresRepository) ListFiguresByDiagram(
	ctx context.Context,
	diagramID string,
) ([]domain.Figure, error) {
	return nil, domain.ErrNotImplemented
}
