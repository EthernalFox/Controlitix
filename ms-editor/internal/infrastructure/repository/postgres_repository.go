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

func (repository *PostgresRepository) CreateTag(
	ctx context.Context,
	tag domain.Tag,
) (domain.Tag, error) {
	query := `
WITH inserted AS (
    INSERT INTO tags.tags (device_id, name, description)
    SELECT d.id, $2, $3
    FROM devices.devices d
    WHERE d.id = $1
      AND d.deleted_at IS NULL
    RETURNING id, device_id, name, description, deleted_at, created_at, updated_at
)
SELECT id, device_id, name, description, deleted_at, created_at, updated_at
FROM inserted
`

	createdTag, queryError := scanTag(
		repository.executor(ctx).QueryRowContext(
			ctx,
			query,
			tag.DeviceID,
			tag.Name,
			tag.Description,
		),
	)
	if queryError != nil {
		return domain.Tag{}, mapDatabaseError("create tag", queryError)
	}

	return createdTag, nil
}

func (repository *PostgresRepository) GetTag(
	ctx context.Context,
	tagID string,
) (domain.TagFull, error) {
	query := `
SELECT
    t.id,
    t.device_id,
    t.name,
    t.description,
    t.deleted_at,
    t.created_at,
    t.updated_at,
    tp.id,
    tp.tag_id,
    tp.data_type_id,
    tp.unit_id,
    tp.address,
    tp.created_at,
    tp.updated_at,
    ts.param_id,
    ts.lolo,
    ts.lo,
    ts.hi,
    ts.hihi,
    ts.created_at,
    ts.updated_at,
    sc.param_id,
    sc.raw_min,
    sc.raw_max,
    sc.eng_min,
    sc.eng_max,
    sc.factor,
    sc."offset",
    sc.created_at,
    sc.updated_at
FROM tags.tags t
LEFT JOIN tags.tag_params tp ON tp.tag_id = t.id
LEFT JOIN tags.tag_setpoints ts ON ts.param_id = tp.id
LEFT JOIN tags.tag_scaling sc ON sc.param_id = tp.id
WHERE t.id = $1
  AND t.deleted_at IS NULL
`

	tagFull, queryError := scanTagFull(
		repository.executor(ctx).QueryRowContext(ctx, query, tagID),
	)
	if queryError != nil {
		return domain.TagFull{}, mapDatabaseError("get tag", queryError)
	}

	return tagFull, nil
}

func (repository *PostgresRepository) UpdateTag(
	ctx context.Context,
	tagID string,
	update domain.TagUpdate,
) (domain.Tag, error) {
	query := `
UPDATE tags.tags
SET
    name = COALESCE($2, name),
    description = COALESCE($3, description),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
RETURNING id, device_id, name, description, deleted_at, created_at, updated_at
`

	tag, queryError := scanTag(
		repository.executor(ctx).QueryRowContext(
			ctx,
			query,
			tagID,
			update.Name,
			update.Description,
		),
	)
	if queryError != nil {
		return domain.Tag{}, mapDatabaseError("update tag", queryError)
	}

	return tag, nil
}

func (repository *PostgresRepository) DeleteTag(
	ctx context.Context,
	tagID string,
) error {
	query := `
UPDATE tags.tags
SET
    deleted_at = now(),
    updated_at = now()
WHERE id = $1
  AND deleted_at IS NULL
`

	result, executeError := repository.executor(ctx).ExecContext(ctx, query, tagID)
	if executeError != nil {
		return mapDatabaseError("delete tag", executeError)
	}

	return ensureRowsAffected("delete tag", result)
}

func (repository *PostgresRepository) ListTagsByDevice(
	ctx context.Context,
	deviceID string,
) ([]domain.Tag, error) {
	query := `
SELECT id, device_id, name, description, deleted_at, created_at, updated_at
FROM tags.tags
WHERE device_id = $1
  AND deleted_at IS NULL
ORDER BY created_at, id
`

	rows, queryError := repository.executor(ctx).QueryContext(ctx, query, deviceID)
	if queryError != nil {
		return nil, mapDatabaseError("list tags by device", queryError)
	}
	defer rows.Close()

	tags := make([]domain.Tag, 0)
	for rows.Next() {
		tag, scanError := scanTag(rows)
		if scanError != nil {
			return nil, mapDatabaseError("list tags by device", scanError)
		}

		tags = append(tags, tag)
	}

	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("list tags by device: %w", rowsError)
	}

	return tags, nil
}

func (repository *PostgresRepository) UpsertTagParams(
	ctx context.Context,
	tagID string,
	params domain.TagParams,
) (domain.TagParams, error) {
	query := `
INSERT INTO tags.tag_params (tag_id, data_type_id, unit_id, address)
SELECT t.id, $2, $3, $4
FROM tags.tags t
WHERE t.id = $1
  AND t.deleted_at IS NULL
ON CONFLICT (tag_id) DO UPDATE
SET
    data_type_id = EXCLUDED.data_type_id,
    unit_id = EXCLUDED.unit_id,
    address = EXCLUDED.address,
    updated_at = now()
RETURNING id, tag_id, data_type_id, unit_id, address, created_at, updated_at
`

	tagParams, queryError := scanTagParams(
		repository.executor(ctx).QueryRowContext(
			ctx,
			query,
			tagID,
			params.DataTypeID,
			params.UnitID,
			[]byte(params.Address),
		),
	)
	if queryError != nil {
		return domain.TagParams{}, mapDatabaseError("upsert tag params", queryError)
	}

	return tagParams, nil
}

func (repository *PostgresRepository) GetTagParams(
	ctx context.Context,
	tagID string,
) (domain.TagParams, error) {
	query := `
SELECT tp.id, tp.tag_id, tp.data_type_id, tp.unit_id, tp.address, tp.created_at, tp.updated_at
FROM tags.tag_params tp
JOIN tags.tags t ON t.id = tp.tag_id
WHERE tp.tag_id = $1
  AND t.deleted_at IS NULL
`

	tagParams, queryError := scanTagParams(
		repository.executor(ctx).QueryRowContext(ctx, query, tagID),
	)
	if queryError != nil {
		return domain.TagParams{}, mapDatabaseError("get tag params", queryError)
	}

	return tagParams, nil
}

func (repository *PostgresRepository) UpsertTagSetpoints(
	ctx context.Context,
	paramID string,
	setpoints domain.TagSetpoints,
) (domain.TagSetpoints, error) {
	query := `
INSERT INTO tags.tag_setpoints (param_id, lolo, lo, hi, hihi)
SELECT tp.id, $2, $3, $4, $5
FROM tags.tag_params tp
JOIN tags.tags t ON t.id = tp.tag_id
WHERE tp.id = $1
  AND t.deleted_at IS NULL
ON CONFLICT (param_id) DO UPDATE
SET
    lolo = EXCLUDED.lolo,
    lo = EXCLUDED.lo,
    hi = EXCLUDED.hi,
    hihi = EXCLUDED.hihi,
    updated_at = now()
RETURNING param_id, lolo, lo, hi, hihi, created_at, updated_at
`

	tagSetpoints, queryError := scanTagSetpoints(
		repository.executor(ctx).QueryRowContext(
			ctx,
			query,
			paramID,
			setpoints.LoLo,
			setpoints.Lo,
			setpoints.Hi,
			setpoints.HiHi,
		),
	)
	if queryError != nil {
		return domain.TagSetpoints{}, mapDatabaseError("upsert tag setpoints", queryError)
	}

	return tagSetpoints, nil
}

func (repository *PostgresRepository) GetTagSetpoints(
	ctx context.Context,
	paramID string,
) (domain.TagSetpoints, error) {
	query := `
SELECT param_id, lolo, lo, hi, hihi, created_at, updated_at
FROM tags.tag_setpoints
WHERE param_id = $1
`

	tagSetpoints, queryError := scanTagSetpoints(
		repository.executor(ctx).QueryRowContext(ctx, query, paramID),
	)
	if queryError != nil {
		return domain.TagSetpoints{}, mapDatabaseError("get tag setpoints", queryError)
	}

	return tagSetpoints, nil
}

func (repository *PostgresRepository) DeleteTagSetpoints(
	ctx context.Context,
	paramID string,
) error {
	result, executeError := repository.executor(ctx).ExecContext(
		ctx,
		`DELETE FROM tags.tag_setpoints WHERE param_id = $1`,
		paramID,
	)
	if executeError != nil {
		return mapDatabaseError("delete tag setpoints", executeError)
	}

	return ensureRowsAffected("delete tag setpoints", result)
}

func (repository *PostgresRepository) UpsertTagScaling(
	ctx context.Context,
	paramID string,
	scaling domain.TagScaling,
) (domain.TagScaling, error) {
	query := `
INSERT INTO tags.tag_scaling (param_id, raw_min, raw_max, eng_min, eng_max, factor, "offset")
SELECT tp.id, $2, $3, $4, $5, $6, $7
FROM tags.tag_params tp
JOIN tags.tags t ON t.id = tp.tag_id
WHERE tp.id = $1
  AND t.deleted_at IS NULL
ON CONFLICT (param_id) DO UPDATE
SET
    raw_min = EXCLUDED.raw_min,
    raw_max = EXCLUDED.raw_max,
    eng_min = EXCLUDED.eng_min,
    eng_max = EXCLUDED.eng_max,
    factor = EXCLUDED.factor,
    "offset" = EXCLUDED."offset",
    updated_at = now()
RETURNING param_id, raw_min, raw_max, eng_min, eng_max, factor, "offset", created_at, updated_at
`

	tagScaling, queryError := scanTagScaling(
		repository.executor(ctx).QueryRowContext(
			ctx,
			query,
			paramID,
			scaling.RawMin,
			scaling.RawMax,
			scaling.EngMin,
			scaling.EngMax,
			scaling.Factor,
			scaling.Offset,
		),
	)
	if queryError != nil {
		return domain.TagScaling{}, mapDatabaseError("upsert tag scaling", queryError)
	}

	return tagScaling, nil
}

func (repository *PostgresRepository) GetTagScaling(
	ctx context.Context,
	paramID string,
) (domain.TagScaling, error) {
	query := `
SELECT param_id, raw_min, raw_max, eng_min, eng_max, factor, "offset", created_at, updated_at
FROM tags.tag_scaling
WHERE param_id = $1
`

	tagScaling, queryError := scanTagScaling(
		repository.executor(ctx).QueryRowContext(ctx, query, paramID),
	)
	if queryError != nil {
		return domain.TagScaling{}, mapDatabaseError("get tag scaling", queryError)
	}

	return tagScaling, nil
}

func (repository *PostgresRepository) DeleteTagScaling(
	ctx context.Context,
	paramID string,
) error {
	result, executeError := repository.executor(ctx).ExecContext(
		ctx,
		`DELETE FROM tags.tag_scaling WHERE param_id = $1`,
		paramID,
	)
	if executeError != nil {
		return mapDatabaseError("delete tag scaling", executeError)
	}

	return ensureRowsAffected("delete tag scaling", result)
}

func (repository *PostgresRepository) ListDataTypes(
	ctx context.Context,
) ([]domain.DataType, error) {
	rows, queryError := repository.executor(ctx).QueryContext(
		ctx,
		`SELECT id, name FROM tags.data_types ORDER BY id`,
	)
	if queryError != nil {
		return nil, mapDatabaseError("list data types", queryError)
	}
	defer rows.Close()

	dataTypes := make([]domain.DataType, 0)
	for rows.Next() {
		var dataType domain.DataType
		if scanError := rows.Scan(&dataType.ID, &dataType.Name); scanError != nil {
			return nil, fmt.Errorf("scan data type: %w", scanError)
		}

		dataTypes = append(dataTypes, dataType)
	}

	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("list data types: %w", rowsError)
	}

	return dataTypes, nil
}

func (repository *PostgresRepository) ListUnits(
	ctx context.Context,
) ([]domain.Unit, error) {
	return repository.listUnitsByQuery(
		ctx,
		`SELECT id, name, symbol, category FROM tags.units ORDER BY id`,
	)
}

func (repository *PostgresRepository) ListUnitsByCategory(
	ctx context.Context,
	category string,
) ([]domain.Unit, error) {
	return repository.listUnitsByQuery(
		ctx,
		`SELECT id, name, symbol, category FROM tags.units WHERE category = $1 ORDER BY id`,
		category,
	)
}

func (repository *PostgresRepository) listUnitsByQuery(
	ctx context.Context,
	query string,
	arguments ...any,
) ([]domain.Unit, error) {
	rows, queryError := repository.executor(ctx).QueryContext(ctx, query, arguments...)
	if queryError != nil {
		return nil, mapDatabaseError("list units", queryError)
	}
	defer rows.Close()

	units := make([]domain.Unit, 0)
	for rows.Next() {
		var unit domain.Unit
		if scanError := rows.Scan(&unit.ID, &unit.Name, &unit.Symbol, &unit.Category); scanError != nil {
			return nil, fmt.Errorf("scan unit: %w", scanError)
		}

		units = append(units, unit)
	}

	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("list units: %w", rowsError)
	}

	return units, nil
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

func scanTag(scanner rowScanner) (domain.Tag, error) {
	var (
		tag         domain.Tag
		description sql.NullString
		deletedAt   sql.NullTime
	)

	scanError := scanner.Scan(
		&tag.ID,
		&tag.DeviceID,
		&tag.Name,
		&description,
		&deletedAt,
		&tag.CreatedAt,
		&tag.UpdatedAt,
	)
	if scanError != nil {
		return domain.Tag{}, scanError
	}

	if description.Valid {
		tag.Description = &description.String
	}

	if deletedAt.Valid {
		tag.DeletedAt = &deletedAt.Time
	}

	return tag, nil
}

func scanTagParams(scanner rowScanner) (domain.TagParams, error) {
	var (
		tagParams domain.TagParams
		unitID    sql.NullInt64
		address   []byte
	)

	scanError := scanner.Scan(
		&tagParams.ID,
		&tagParams.TagID,
		&tagParams.DataTypeID,
		&unitID,
		&address,
		&tagParams.CreatedAt,
		&tagParams.UpdatedAt,
	)
	if scanError != nil {
		return domain.TagParams{}, scanError
	}

	if unitID.Valid {
		tagParams.UnitID = intPointerFromInt64(unitID.Int64)
	}

	if address != nil {
		tagParams.Address = append(json.RawMessage(nil), address...)
	}

	return tagParams, nil
}

func scanTagSetpoints(scanner rowScanner) (domain.TagSetpoints, error) {
	var (
		tagSetpoints domain.TagSetpoints
		lolo         sql.NullFloat64
		lo           sql.NullFloat64
		hi           sql.NullFloat64
		hihi         sql.NullFloat64
	)

	scanError := scanner.Scan(
		&tagSetpoints.ParamID,
		&lolo,
		&lo,
		&hi,
		&hihi,
		&tagSetpoints.CreatedAt,
		&tagSetpoints.UpdatedAt,
	)
	if scanError != nil {
		return domain.TagSetpoints{}, scanError
	}

	tagSetpoints.LoLo = floatPointerFromNull(lolo)
	tagSetpoints.Lo = floatPointerFromNull(lo)
	tagSetpoints.Hi = floatPointerFromNull(hi)
	tagSetpoints.HiHi = floatPointerFromNull(hihi)

	return tagSetpoints, nil
}

func scanTagScaling(scanner rowScanner) (domain.TagScaling, error) {
	var (
		tagScaling domain.TagScaling
		rawMin     sql.NullFloat64
		rawMax     sql.NullFloat64
		engMin     sql.NullFloat64
		engMax     sql.NullFloat64
		factor     sql.NullFloat64
		offset     sql.NullFloat64
	)

	scanError := scanner.Scan(
		&tagScaling.ParamID,
		&rawMin,
		&rawMax,
		&engMin,
		&engMax,
		&factor,
		&offset,
		&tagScaling.CreatedAt,
		&tagScaling.UpdatedAt,
	)
	if scanError != nil {
		return domain.TagScaling{}, scanError
	}

	tagScaling.RawMin = floatPointerFromNull(rawMin)
	tagScaling.RawMax = floatPointerFromNull(rawMax)
	tagScaling.EngMin = floatPointerFromNull(engMin)
	tagScaling.EngMax = floatPointerFromNull(engMax)
	tagScaling.Factor = floatPointerFromNull(factor)
	tagScaling.Offset = floatPointerFromNull(offset)

	return tagScaling, nil
}

func scanTagFull(scanner rowScanner) (domain.TagFull, error) {
	var (
		tagFull          domain.TagFull
		description      sql.NullString
		deletedAt        sql.NullTime
		tagParamsID      sql.NullString
		tagParamsTagID   sql.NullString
		dataTypeID       sql.NullInt64
		unitID           sql.NullInt64
		address          []byte
		paramsCreatedAt  sql.NullTime
		paramsUpdatedAt  sql.NullTime
		setpointsParamID sql.NullString
		lolo             sql.NullFloat64
		lo               sql.NullFloat64
		hi               sql.NullFloat64
		hihi             sql.NullFloat64
		setCreatedAt     sql.NullTime
		setUpdatedAt     sql.NullTime
		scalingParamID   sql.NullString
		rawMin           sql.NullFloat64
		rawMax           sql.NullFloat64
		engMin           sql.NullFloat64
		engMax           sql.NullFloat64
		factor           sql.NullFloat64
		offset           sql.NullFloat64
		scalingCreatedAt sql.NullTime
		scalingUpdatedAt sql.NullTime
	)

	scanError := scanner.Scan(
		&tagFull.Tag.ID,
		&tagFull.Tag.DeviceID,
		&tagFull.Tag.Name,
		&description,
		&deletedAt,
		&tagFull.Tag.CreatedAt,
		&tagFull.Tag.UpdatedAt,
		&tagParamsID,
		&tagParamsTagID,
		&dataTypeID,
		&unitID,
		&address,
		&paramsCreatedAt,
		&paramsUpdatedAt,
		&setpointsParamID,
		&lolo,
		&lo,
		&hi,
		&hihi,
		&setCreatedAt,
		&setUpdatedAt,
		&scalingParamID,
		&rawMin,
		&rawMax,
		&engMin,
		&engMax,
		&factor,
		&offset,
		&scalingCreatedAt,
		&scalingUpdatedAt,
	)
	if scanError != nil {
		return domain.TagFull{}, scanError
	}

	if description.Valid {
		tagFull.Tag.Description = &description.String
	}

	if deletedAt.Valid {
		tagFull.Tag.DeletedAt = &deletedAt.Time
	}

	if tagParamsID.Valid {
		tagFull.Params = &domain.TagParams{
			ID:         tagParamsID.String,
			TagID:      tagParamsTagID.String,
			DataTypeID: int(dataTypeID.Int64),
			UnitID:     intPointerFromInt64(unitID.Int64),
			CreatedAt:  paramsCreatedAt.Time,
			UpdatedAt:  paramsUpdatedAt.Time,
		}
		if address != nil {
			tagFull.Params.Address = append(json.RawMessage(nil), address...)
		}
		if !unitID.Valid {
			tagFull.Params.UnitID = nil
		}
	}

	if setpointsParamID.Valid {
		tagFull.Setpoints = &domain.TagSetpoints{
			ParamID:   setpointsParamID.String,
			LoLo:      floatPointerFromNull(lolo),
			Lo:        floatPointerFromNull(lo),
			Hi:        floatPointerFromNull(hi),
			HiHi:      floatPointerFromNull(hihi),
			CreatedAt: setCreatedAt.Time,
			UpdatedAt: setUpdatedAt.Time,
		}
	}

	if scalingParamID.Valid {
		tagFull.Scaling = &domain.TagScaling{
			ParamID:   scalingParamID.String,
			RawMin:    floatPointerFromNull(rawMin),
			RawMax:    floatPointerFromNull(rawMax),
			EngMin:    floatPointerFromNull(engMin),
			EngMax:    floatPointerFromNull(engMax),
			Factor:    floatPointerFromNull(factor),
			Offset:    floatPointerFromNull(offset),
			CreatedAt: scalingCreatedAt.Time,
			UpdatedAt: scalingUpdatedAt.Time,
		}
	}

	return tagFull, nil
}

func intPointerFromInt64(value int64) *int {
	convertedValue := int(value)
	return &convertedValue
}

func floatPointerFromNull(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}

	return &value.Float64
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
var _ domain.TagRepository = (*PostgresRepository)(nil)
var _ domain.TagParamsRepository = (*PostgresRepository)(nil)
var _ domain.TagSetpointsRepository = (*PostgresRepository)(nil)
var _ domain.TagScalingRepository = (*PostgresRepository)(nil)
var _ domain.ReferenceRepository = (*PostgresRepository)(nil)
var _ interface {
	InTransaction(ctx context.Context, operation func(context.Context) error) error
} = (*PostgresRepository)(nil)
