package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/EthernalFox/Controlitix/ms-poll/internal/domain"
)

const loadAllDevicesQuery = `
SELECT d.id, d.object_id, d.type_id, dt.name, d.name,
       COALESCE(p.settings, '{}'::jsonb) AS settings,
       d.updated_at
  FROM devices.devices d
  JOIN devices.device_type dt ON dt.id = d.type_id
  LEFT JOIN devices.devices_params p ON p.device_id = d.id
 WHERE d.deleted_at IS NULL;
`

const loadAllTagsQueryTemplate = `
SELECT t.id, t.device_id, t.name,
       dty.name AS data_type, u.symbol AS unit_symbol,
       tp.address,
       sc.raw_min, sc.raw_max, sc.eng_min, sc.eng_max, sc.factor, sc."offset",
       sp.lolo, sp.lo, sp.hi, sp.hihi,
       t.updated_at
  FROM tags.tags t
  JOIN tags.tag_params tp ON tp.tag_id = t.id
  JOIN tags.data_types dty ON dty.id = tp.data_type_id
  LEFT JOIN tags.units u ON u.id = tp.unit_id
  LEFT JOIN tags.tag_scaling sc ON sc.param_id = tp.id
  LEFT JOIN tags.tag_setpoints sp ON sp.param_id = tp.id
 WHERE t.device_id IN (%s) AND t.deleted_at IS NULL;
`

const loadDeviceByIDQuery = `
SELECT d.id, d.object_id, d.type_id, dt.name, d.name,
       COALESCE(p.settings, '{}'::jsonb) AS settings,
       d.updated_at
  FROM devices.devices d
  JOIN devices.device_type dt ON dt.id = d.type_id
  LEFT JOIN devices.devices_params p ON p.device_id = d.id
 WHERE d.id = $1 AND d.deleted_at IS NULL;
`

const loadTagByIDQuery = `
SELECT t.id, t.device_id, t.name,
       dty.name AS data_type, u.symbol AS unit_symbol,
       tp.address,
       sc.raw_min, sc.raw_max, sc.eng_min, sc.eng_max, sc.factor, sc."offset",
       sp.lolo, sp.lo, sp.hi, sp.hihi,
       t.updated_at
  FROM tags.tags t
  JOIN tags.tag_params tp ON tp.tag_id = t.id
  JOIN tags.data_types dty ON dty.id = tp.data_type_id
  LEFT JOIN tags.units u ON u.id = tp.unit_id
  LEFT JOIN tags.tag_scaling sc ON sc.param_id = tp.id
  LEFT JOIN tags.tag_setpoints sp ON sp.param_id = tp.id
 WHERE t.id = $1 AND t.deleted_at IS NULL;
`

type ConfigRepository struct {
	databaseConnection *sql.DB
}

func NewConfigRepository(databaseConnection *sql.DB) *ConfigRepository {
	return &ConfigRepository{
		databaseConnection: databaseConnection,
	}
}

func (repository *ConfigRepository) LoadAll(
	ctx context.Context,
) (*domain.Snapshot, error) {
	deviceRows, queryDeviceError := repository.databaseConnection.QueryContext(
		ctx,
		loadAllDevicesQuery,
	)
	if queryDeviceError != nil {
		return nil, fmt.Errorf("query devices: %w", queryDeviceError)
	}
	defer deviceRows.Close()

	snapshot := &domain.Snapshot{
		Devices:      make(map[string]*domain.DeviceSnapshot),
		Tags:         make(map[string]*domain.TagSnapshot),
		TagsByDevice: make(map[string][]*domain.TagSnapshot),
	}

	deviceIDs := make([]string, 0)
	for deviceRows.Next() {
		var objectID sql.NullString
		var settings json.RawMessage

		deviceSnapshot := &domain.DeviceSnapshot{}
		scanDeviceError := deviceRows.Scan(
			&deviceSnapshot.ID,
			&objectID,
			&deviceSnapshot.TypeID,
			&deviceSnapshot.TypeName,
			&deviceSnapshot.Name,
			&settings,
			&deviceSnapshot.UpdatedAt,
		)
		if scanDeviceError != nil {
			return nil, fmt.Errorf("scan device: %w", scanDeviceError)
		}

		deviceSnapshot.ObjectID = nullableStringToPointer(objectID)
		deviceSnapshot.Settings = cloneRawJSON(settings)

		snapshot.Devices[deviceSnapshot.ID] = deviceSnapshot
		deviceIDs = append(deviceIDs, deviceSnapshot.ID)
	}

	if iterDeviceError := deviceRows.Err(); iterDeviceError != nil {
		return nil, fmt.Errorf("iterate devices rows: %w", iterDeviceError)
	}

	if len(deviceIDs) == 0 {
		return snapshot, nil
	}

	tagQuery, tagQueryArguments := buildLoadAllTagsQuery(deviceIDs)
	tagRows, queryTagError := repository.databaseConnection.QueryContext(
		ctx,
		tagQuery,
		tagQueryArguments...,
	)
	if queryTagError != nil {
		return nil, fmt.Errorf("query tags: %w", queryTagError)
	}
	defer tagRows.Close()

	for tagRows.Next() {
		tagSnapshot, scanTagError := scanTagSnapshotRow(tagRows)
		if scanTagError != nil {
			return nil, fmt.Errorf("scan tag: %w", scanTagError)
		}

		snapshot.Tags[tagSnapshot.ID] = tagSnapshot
		snapshot.TagsByDevice[tagSnapshot.DeviceID] = append(
			snapshot.TagsByDevice[tagSnapshot.DeviceID],
			tagSnapshot,
		)
	}

	if iterTagError := tagRows.Err(); iterTagError != nil {
		return nil, fmt.Errorf("iterate tags rows: %w", iterTagError)
	}

	return snapshot, nil
}

func (repository *ConfigRepository) LoadDevice(
	ctx context.Context,
	id string,
) (*domain.DeviceSnapshot, error) {
	var objectID sql.NullString
	var settings json.RawMessage
	deviceSnapshot := &domain.DeviceSnapshot{}

	queryError := repository.databaseConnection.QueryRowContext(
		ctx,
		loadDeviceByIDQuery,
		id,
	).Scan(
		&deviceSnapshot.ID,
		&objectID,
		&deviceSnapshot.TypeID,
		&deviceSnapshot.TypeName,
		&deviceSnapshot.Name,
		&settings,
		&deviceSnapshot.UpdatedAt,
	)
	if queryError != nil {
		if errors.Is(queryError, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query device by id: %w", queryError)
	}

	deviceSnapshot.ObjectID = nullableStringToPointer(objectID)
	deviceSnapshot.Settings = cloneRawJSON(settings)

	return deviceSnapshot, nil
}

func (repository *ConfigRepository) LoadTag(
	ctx context.Context,
	id string,
) (*domain.TagSnapshot, error) {
	row := repository.databaseConnection.QueryRowContext(ctx, loadTagByIDQuery, id)
	tagSnapshot, scanError := scanTagSnapshotRow(row)
	if scanError != nil {
		if errors.Is(scanError, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("query tag by id: %w", scanError)
	}

	return tagSnapshot, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanTagSnapshotRow(row scanner) (*domain.TagSnapshot, error) {
	var unitSymbol sql.NullString
	var address json.RawMessage

	var rawMin sql.NullFloat64
	var rawMax sql.NullFloat64
	var engMin sql.NullFloat64
	var engMax sql.NullFloat64
	var factor sql.NullFloat64
	var offset sql.NullFloat64

	var lolo sql.NullFloat64
	var lo sql.NullFloat64
	var hi sql.NullFloat64
	var hihi sql.NullFloat64

	tagSnapshot := &domain.TagSnapshot{}
	scanError := row.Scan(
		&tagSnapshot.ID,
		&tagSnapshot.DeviceID,
		&tagSnapshot.Name,
		&tagSnapshot.DataType,
		&unitSymbol,
		&address,
		&rawMin,
		&rawMax,
		&engMin,
		&engMax,
		&factor,
		&offset,
		&lolo,
		&lo,
		&hi,
		&hihi,
		&tagSnapshot.UpdatedAt,
	)
	if scanError != nil {
		return nil, scanError
	}

	tagSnapshot.UnitSymbol = nullableStringToPointer(unitSymbol)
	tagSnapshot.Address = cloneRawJSON(address)
	tagSnapshot.Scaling = nullableTagScaling(
		rawMin,
		rawMax,
		engMin,
		engMax,
		factor,
		offset,
	)
	tagSnapshot.Setpoints = nullableTagSetpoints(lolo, lo, hi, hihi)

	return tagSnapshot, nil
}

func nullableStringToPointer(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	copiedValue := value.String
	return &copiedValue
}

func nullableTagScaling(
	rawMin sql.NullFloat64,
	rawMax sql.NullFloat64,
	engMin sql.NullFloat64,
	engMax sql.NullFloat64,
	factor sql.NullFloat64,
	offset sql.NullFloat64,
) *domain.TagScaling {
	tagScaling := &domain.TagScaling{
		RawMin: nullableFloatToPointer(rawMin),
		RawMax: nullableFloatToPointer(rawMax),
		EngMin: nullableFloatToPointer(engMin),
		EngMax: nullableFloatToPointer(engMax),
		Factor: nullableFloatToPointer(factor),
		Offset: nullableFloatToPointer(offset),
	}

	if tagScaling.RawMin == nil &&
		tagScaling.RawMax == nil &&
		tagScaling.EngMin == nil &&
		tagScaling.EngMax == nil &&
		tagScaling.Factor == nil &&
		tagScaling.Offset == nil {
		return nil
	}

	return tagScaling
}

func nullableTagSetpoints(
	lolo sql.NullFloat64,
	lo sql.NullFloat64,
	hi sql.NullFloat64,
	hihi sql.NullFloat64,
) *domain.TagSetpoints {
	tagSetpoints := &domain.TagSetpoints{
		LoLo: nullableFloatToPointer(lolo),
		Lo:   nullableFloatToPointer(lo),
		Hi:   nullableFloatToPointer(hi),
		HiHi: nullableFloatToPointer(hihi),
	}

	if tagSetpoints.LoLo == nil &&
		tagSetpoints.Lo == nil &&
		tagSetpoints.Hi == nil &&
		tagSetpoints.HiHi == nil {
		return nil
	}

	return tagSetpoints
}

func nullableFloatToPointer(value sql.NullFloat64) *float64 {
	if !value.Valid {
		return nil
	}

	copiedValue := value.Float64
	return &copiedValue
}

func cloneRawJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return nil
	}

	return append(json.RawMessage(nil), raw...)
}

func buildLoadAllTagsQuery(deviceIDs []string) (string, []any) {
	placeholders := make([]string, 0, len(deviceIDs))
	arguments := make([]any, 0, len(deviceIDs))
	for index, deviceID := range deviceIDs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+1))
		arguments = append(arguments, deviceID)
	}

	return fmt.Sprintf(
		loadAllTagsQueryTemplate,
		strings.Join(placeholders, ", "),
	), arguments
}
