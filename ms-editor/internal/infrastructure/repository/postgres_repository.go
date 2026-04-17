package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/EthernalFox/Controlitix/ms-editor/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
)

type transactionContextKey string

const postgresTransactionContextKey transactionContextKey = "postgres_transaction"

type sqlExecutor interface {
	ExecContext(ctx context.Context, query string, arguments ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, arguments ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, arguments ...any) *sql.Row
}

type rowScanner interface {
	Scan(destinations ...any) error
}

type PostgresRepository struct {
	databaseConnection *sql.DB
}

func NewPostgresRepository(databaseConnection *sql.DB) *PostgresRepository {
	return &PostgresRepository{
		databaseConnection: databaseConnection,
	}
}

func (repository *PostgresRepository) InTransaction(
	ctx context.Context,
	operation func(context.Context) error,
) error {
	transaction, beginError := repository.databaseConnection.BeginTx(ctx, nil)
	if beginError != nil {
		return fmt.Errorf("begin transaction: %w", beginError)
	}

	transactionContext := context.WithValue(
		ctx,
		postgresTransactionContextKey,
		transaction,
	)

	if operationError := operation(transactionContext); operationError != nil {
		if rollbackError := transaction.Rollback(); rollbackError != nil {
			return errors.Join(operationError, fmt.Errorf("rollback transaction: %w", rollbackError))
		}

		return operationError
	}

	if commitError := transaction.Commit(); commitError != nil {
		return fmt.Errorf("commit transaction: %w", commitError)
	}

	return nil
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

func (repository *PostgresRepository) CreateDevice(
	ctx context.Context,
	device domain.Device,
) (domain.Device, error) {
	query := `
WITH inserted AS (
    INSERT INTO devices.devices (object_id, type_id, name, description)
    VALUES ($1, $2, $3, $4)
    RETURNING id, object_id, type_id, name, description, deleted_at, created_at, updated_at
)
SELECT
    inserted.id,
    inserted.object_id,
    inserted.type_id,
    device_type.name,
    inserted.name,
    inserted.description,
    inserted.deleted_at,
    inserted.created_at,
    inserted.updated_at
FROM inserted
JOIN devices.device_type ON device_type.id = inserted.type_id
`

	createdDevice, queryError := scanDevice(
		repository.executor(ctx).QueryRowContext(
			ctx,
			query,
			device.ObjectID,
			device.TypeID,
			device.Name,
			device.Description,
		),
	)
	if queryError != nil {
		return domain.Device{}, mapDatabaseError("create device", queryError)
	}

	return createdDevice, nil
}

func (repository *PostgresRepository) GetDevice(
	ctx context.Context,
	deviceID string,
) (domain.DeviceWithParams, error) {
	query := `
SELECT
    d.id,
    d.object_id,
    d.type_id,
    dt.name,
    d.name,
    d.description,
    d.deleted_at,
    d.created_at,
    d.updated_at,
    dp.settings,
    dp.created_at,
    dp.updated_at
FROM devices.devices d
JOIN devices.device_type dt ON dt.id = d.type_id
LEFT JOIN devices.devices_params dp ON dp.device_id = d.id
WHERE d.id = $1
  AND d.deleted_at IS NULL
`

	deviceWithParams, queryError := scanDeviceWithParams(
		repository.executor(ctx).QueryRowContext(ctx, query, deviceID),
	)
	if queryError != nil {
		return domain.DeviceWithParams{}, mapDatabaseError("get device", queryError)
	}

	return deviceWithParams, nil
}

func (repository *PostgresRepository) UpdateDevice(
	ctx context.Context,
	deviceID string,
	update domain.DeviceUpdate,
) (domain.Device, error) {
	query := `
UPDATE devices.devices
SET
    type_id = COALESCE($2, type_id),
    name = COALESCE($3, name),
    description = COALESCE($4, description),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING
    id,
    object_id,
    type_id,
    (
        SELECT name
        FROM devices.device_type
        WHERE id = devices.devices.type_id
    ),
    name,
    description,
    deleted_at,
    created_at,
    updated_at
`

	device, queryError := scanDevice(
		repository.executor(ctx).QueryRowContext(
			ctx,
			query,
			deviceID,
			update.TypeID,
			update.Name,
			update.Description,
		),
	)
	if queryError != nil {
		return domain.Device{}, mapDatabaseError("update device", queryError)
	}

	return device, nil
}

func (repository *PostgresRepository) DeleteDevice(
	ctx context.Context,
	deviceID string,
) error {
	query := `
UPDATE devices.devices
SET
    deleted_at = now(),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
`

	result, executeError := repository.executor(ctx).ExecContext(ctx, query, deviceID)
	if executeError != nil {
		return mapDatabaseError("delete device", executeError)
	}

	return ensureRowsAffected("delete device", result)
}

func (repository *PostgresRepository) ListDevicesByObject(
	ctx context.Context,
	objectID string,
) ([]domain.Device, error) {
	query := `
SELECT
    d.id,
    d.object_id,
    d.type_id,
    dt.name,
    d.name,
    d.description,
    d.deleted_at,
    d.created_at,
    d.updated_at
FROM devices.devices d
JOIN devices.device_type dt ON dt.id = d.type_id
WHERE d.object_id = $1
  AND d.deleted_at IS NULL
ORDER BY d.created_at, d.id
`

	rows, queryError := repository.executor(ctx).QueryContext(ctx, query, objectID)
	if queryError != nil {
		return nil, mapDatabaseError("list devices by object", queryError)
	}
	defer rows.Close()

	devices := make([]domain.Device, 0)
	for rows.Next() {
		device, scanError := scanDevice(rows)
		if scanError != nil {
			return nil, mapDatabaseError("list devices by object", scanError)
		}

		devices = append(devices, device)
	}

	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("list devices by object: %w", rowsError)
	}

	return devices, nil
}

func (repository *PostgresRepository) AssignDeviceToObject(
	ctx context.Context,
	deviceID string,
	objectID *string,
) (domain.Device, error) {
	query := `
UPDATE devices.devices
SET
    object_id = $2,
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING
    id,
    object_id,
    type_id,
    (
        SELECT name
        FROM devices.device_type
        WHERE id = devices.devices.type_id
    ),
    name,
    description,
    deleted_at,
    created_at,
    updated_at
`

	device, queryError := scanDevice(
		repository.executor(ctx).QueryRowContext(ctx, query, deviceID, objectID),
	)
	if queryError != nil {
		return domain.Device{}, mapDatabaseError("assign device to object", queryError)
	}

	return device, nil
}

func (repository *PostgresRepository) UpsertDeviceParams(
	ctx context.Context,
	deviceID string,
	settings json.RawMessage,
) (domain.DeviceParams, error) {
	query := `
INSERT INTO devices.devices_params (device_id, settings)
SELECT d.id, $2
FROM devices.devices d
WHERE d.id = $1
  AND d.deleted_at IS NULL
ON CONFLICT (device_id) DO UPDATE
SET
    settings = EXCLUDED.settings,
    updated_at = now()
RETURNING device_id, settings, created_at, updated_at
`

	deviceParams, queryError := scanDeviceParams(
		repository.executor(ctx).QueryRowContext(ctx, query, deviceID, []byte(settings)),
	)
	if queryError != nil {
		return domain.DeviceParams{}, mapDatabaseError("upsert device params", queryError)
	}

	return deviceParams, nil
}

func (repository *PostgresRepository) GetDeviceParams(
	ctx context.Context,
	deviceID string,
) (domain.DeviceParams, error) {
	query := `
SELECT dp.device_id, dp.settings, dp.created_at, dp.updated_at
FROM devices.devices_params dp
JOIN devices.devices d ON d.id = dp.device_id
WHERE dp.device_id = $1
  AND d.deleted_at IS NULL
`

	deviceParams, queryError := scanDeviceParams(
		repository.executor(ctx).QueryRowContext(ctx, query, deviceID),
	)
	if queryError != nil {
		return domain.DeviceParams{}, mapDatabaseError("get device params", queryError)
	}

	return deviceParams, nil
}

func (repository *PostgresRepository) ListDeviceTypes(
	ctx context.Context,
) ([]domain.DeviceType, error) {
	rows, queryError := repository.executor(ctx).QueryContext(
		ctx,
		`SELECT id, name FROM devices.device_type ORDER BY id`,
	)
	if queryError != nil {
		return nil, mapDatabaseError("list device types", queryError)
	}
	defer rows.Close()

	deviceTypes := make([]domain.DeviceType, 0)
	for rows.Next() {
		var deviceType domain.DeviceType
		if scanError := rows.Scan(&deviceType.ID, &deviceType.Name); scanError != nil {
			return nil, fmt.Errorf("scan device type: %w", scanError)
		}

		deviceTypes = append(deviceTypes, deviceType)
	}

	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("list device types: %w", rowsError)
	}

	return deviceTypes, nil
}

func (repository *PostgresRepository) GetDeviceType(
	ctx context.Context,
	typeID int,
) (domain.DeviceType, error) {
	var deviceType domain.DeviceType
	queryError := repository.executor(ctx).QueryRowContext(
		ctx,
		`SELECT id, name FROM devices.device_type WHERE id = $1`,
		typeID,
	).Scan(&deviceType.ID, &deviceType.Name)
	if queryError != nil {
		return domain.DeviceType{}, mapDatabaseError("get device type", queryError)
	}

	return deviceType, nil
}

func (repository *PostgresRepository) executor(ctx context.Context) sqlExecutor {
	transaction, ok := ctx.Value(postgresTransactionContextKey).(*sql.Tx)
	if ok {
		return transaction
	}

	return repository.databaseConnection
}

func ensureRowsAffected(operation string, result sql.Result) error {
	rowsAffected, rowsError := result.RowsAffected()
	if rowsError != nil {
		return fmt.Errorf("%s rows affected: %w", operation, rowsError)
	}

	if rowsAffected == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func scanDevice(scanner rowScanner) (domain.Device, error) {
	var (
		device      domain.Device
		objectID    sql.NullString
		description sql.NullString
		deletedAt   sql.NullTime
	)

	scanError := scanner.Scan(
		&device.ID,
		&objectID,
		&device.TypeID,
		&device.TypeName,
		&device.Name,
		&description,
		&deletedAt,
		&device.CreatedAt,
		&device.UpdatedAt,
	)
	if scanError != nil {
		return domain.Device{}, scanError
	}

	if objectID.Valid {
		device.ObjectID = &objectID.String
	}

	if description.Valid {
		device.Description = &description.String
	}

	if deletedAt.Valid {
		device.DeletedAt = &deletedAt.Time
	}

	return device, nil
}

func scanDeviceParams(scanner rowScanner) (domain.DeviceParams, error) {
	var (
		deviceParams domain.DeviceParams
		settings     []byte
	)

	scanError := scanner.Scan(
		&deviceParams.DeviceID,
		&settings,
		&deviceParams.CreatedAt,
		&deviceParams.UpdatedAt,
	)
	if scanError != nil {
		return domain.DeviceParams{}, scanError
	}

	deviceParams.Settings = append(json.RawMessage(nil), settings...)
	return deviceParams, nil
}

func scanDeviceWithParams(scanner rowScanner) (domain.DeviceWithParams, error) {
	var (
		deviceWithParams domain.DeviceWithParams
		objectID         sql.NullString
		description      sql.NullString
		deletedAt        sql.NullTime
		settings         []byte
		paramsCreatedAt  sql.NullTime
		paramsUpdatedAt  sql.NullTime
	)

	scanError := scanner.Scan(
		&deviceWithParams.Device.ID,
		&objectID,
		&deviceWithParams.Device.TypeID,
		&deviceWithParams.Device.TypeName,
		&deviceWithParams.Device.Name,
		&description,
		&deletedAt,
		&deviceWithParams.Device.CreatedAt,
		&deviceWithParams.Device.UpdatedAt,
		&settings,
		&paramsCreatedAt,
		&paramsUpdatedAt,
	)
	if scanError != nil {
		return domain.DeviceWithParams{}, scanError
	}

	if objectID.Valid {
		deviceWithParams.Device.ObjectID = &objectID.String
	}

	if description.Valid {
		deviceWithParams.Device.Description = &description.String
	}

	if deletedAt.Valid {
		deviceWithParams.Device.DeletedAt = &deletedAt.Time
	}

	if settings != nil {
		deviceWithParams.Params = &domain.DeviceParams{
			DeviceID:  deviceWithParams.Device.ID,
			Settings:  append(json.RawMessage(nil), settings...),
			CreatedAt: paramsCreatedAt.Time,
			UpdatedAt: paramsUpdatedAt.Time,
		}
	}

	return deviceWithParams, nil
}

func mapDatabaseError(operation string, err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrNotFound
	}

	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23503", "23505":
			return domain.ErrConflict
		}
	}

	return fmt.Errorf("%s: %w", operation, err)
}

var _ domain.DeviceRepository = (*PostgresRepository)(nil)
var _ domain.DeviceParamsRepository = (*PostgresRepository)(nil)
var _ domain.DeviceTypeRepository = (*PostgresRepository)(nil)
var _ interface {
	InTransaction(ctx context.Context, operation func(context.Context) error) error
} = (*PostgresRepository)(nil)
