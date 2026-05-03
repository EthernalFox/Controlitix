package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/EthernalFox/Controlitix/ms-viewer/internal/domain"
)

type TelegramChatsRepository struct {
	pool *pgxpool.Pool
}

func NewTelegramChatsRepository(pool *pgxpool.Pool) *TelegramChatsRepository {
	return &TelegramChatsRepository{pool: pool}
}

func (repository *TelegramChatsRepository) ListChats(
	ctx context.Context,
	filter domain.TelegramChatFilter,
) (domain.TelegramChatListResult, error) {
	conditions := []string{"WHERE 1=1"}
	args := make([]any, 0, 3)
	position := 1

	if filter.Enabled != nil {
		conditions = append(conditions, fmt.Sprintf("AND enabled = $%d", position))
		args = append(args, *filter.Enabled)
		position++
	}
	if role := strings.TrimSpace(filter.Role); role != "" {
		conditions = append(conditions, fmt.Sprintf("AND role = $%d", position))
		args = append(args, role)
		position++
	}
	if filter.ObjectID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("AND object_id = $%d", position))
		args = append(args, filter.ObjectID)
	}

	query := `
SELECT
    id,
    chat_id,
    COALESCE(title, ''),
    user_id,
    role,
    object_id,
    severity_min,
    enabled,
    created_at,
    updated_at
FROM auth.telegram_chats
` + strings.Join(conditions, "\n") + `
ORDER BY created_at DESC, id DESC
`

	rows, queryError := repository.pool.Query(ctx, query, args...)
	if queryError != nil {
		return domain.TelegramChatListResult{}, fmt.Errorf("query telegram chats: %w", queryError)
	}
	defer rows.Close()

	items := make([]domain.TelegramChat, 0)
	for rows.Next() {
		chat, scanError := scanTelegramChat(rows)
		if scanError != nil {
			return domain.TelegramChatListResult{}, scanError
		}
		items = append(items, chat)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return domain.TelegramChatListResult{}, fmt.Errorf("iterate telegram chats: %w", rowsError)
	}

	return domain.TelegramChatListResult{Items: items, Total: len(items)}, nil
}

func (repository *TelegramChatsRepository) ListChatsForRouting(
	ctx context.Context,
	severityMin int16,
	objectID uuid.UUID,
) ([]domain.TelegramChat, error) {
	const query = `
SELECT
    id,
    chat_id,
    COALESCE(title, ''),
    user_id,
    role,
    object_id,
    severity_min,
    enabled,
    created_at,
    updated_at
FROM auth.telegram_chats
WHERE enabled = true
  AND severity_min <= $1
  AND (object_id IS NULL OR object_id = $2)
ORDER BY created_at ASC, id ASC
`

	rows, queryError := repository.pool.Query(ctx, query, severityMin, objectID)
	if queryError != nil {
		return nil, fmt.Errorf("query chats for routing: %w", queryError)
	}
	defer rows.Close()

	items := make([]domain.TelegramChat, 0)
	for rows.Next() {
		chat, scanError := scanTelegramChat(rows)
		if scanError != nil {
			return nil, scanError
		}
		items = append(items, chat)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("iterate chats for routing: %w", rowsError)
	}

	return items, nil
}

func (repository *TelegramChatsRepository) ListAdminChatsForObject(
	ctx context.Context,
	objectID uuid.UUID,
) ([]domain.TelegramChat, error) {
	const query = `
SELECT
    id,
    chat_id,
    COALESCE(title, ''),
    user_id,
    role,
    object_id,
    severity_min,
    enabled,
    created_at,
    updated_at
FROM auth.telegram_chats
WHERE enabled = true
  AND role = 'admin'
  AND (object_id IS NULL OR object_id = $1)
ORDER BY created_at ASC, id ASC
`

	rows, queryError := repository.pool.Query(ctx, query, objectID)
	if queryError != nil {
		return nil, fmt.Errorf("query admin chats: %w", queryError)
	}
	defer rows.Close()

	items := make([]domain.TelegramChat, 0)
	for rows.Next() {
		chat, scanError := scanTelegramChat(rows)
		if scanError != nil {
			return nil, scanError
		}
		items = append(items, chat)
	}
	if rowsError := rows.Err(); rowsError != nil {
		return nil, fmt.Errorf("iterate admin chats: %w", rowsError)
	}

	return items, nil
}

func (repository *TelegramChatsRepository) GetAlarmRouteContext(
	ctx context.Context,
	tagID uuid.UUID,
) (domain.AlarmRouteContext, error) {
	const query = `
SELECT
    t.id,
    COALESCE(t.name, ''),
    d.id,
    COALESCE(d.name, ''),
    COALESCE(o.id, '00000000-0000-0000-0000-000000000000')::uuid,
    COALESCE(o.name, ''),
    COALESCE(u.symbol, '')
FROM tags.tags t
JOIN devices.devices d ON d.id = t.device_id
LEFT JOIN public.objects o ON o.id = d.object_id
LEFT JOIN tags.units u ON u.id = t.unit_id
WHERE t.id = $1
  AND t.deleted_at IS NULL
`

	var routeContext domain.AlarmRouteContext
	queryError := repository.pool.QueryRow(ctx, query, tagID).Scan(
		&routeContext.TagID,
		&routeContext.TagName,
		&routeContext.DeviceID,
		&routeContext.DeviceName,
		&routeContext.ObjectID,
		&routeContext.ObjectName,
		&routeContext.UnitSymbol,
	)
	if queryError != nil {
		if queryError == pgx.ErrNoRows {
			return domain.AlarmRouteContext{}, fmt.Errorf("tag %s not found: %w", tagID.String(), domain.ErrNotFound)
		}
		return domain.AlarmRouteContext{}, fmt.Errorf("query route context: %w", queryError)
	}

	return routeContext, nil
}

func scanTelegramChat(rows pgx.Rows) (domain.TelegramChat, error) {
	var chat domain.TelegramChat
	if scanError := rows.Scan(
		&chat.ID,
		&chat.ChatID,
		&chat.Title,
		&chat.UserID,
		&chat.Role,
		&chat.ObjectID,
		&chat.SeverityMin,
		&chat.Enabled,
		&chat.CreatedAt,
		&chat.UpdatedAt,
	); scanError != nil {
		return domain.TelegramChat{}, fmt.Errorf("scan telegram chat: %w", scanError)
	}

	return chat, nil
}
