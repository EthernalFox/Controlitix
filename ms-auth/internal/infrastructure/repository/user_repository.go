package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

type UserListFilter struct {
	Offset   int
	Limit    int
	Query    string
	SourceID *int
	IsActive *bool
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindBySourceAndUsername(
	ctx context.Context,
	sourceID int,
	username string,
) (*domain.User, error) {
	const query = `
SELECT id, subject, source_id, username, display_name, email, password_hash, is_active, last_login_at, created_at, updated_at
FROM auth.users
WHERE source_id = $1 AND lower(username) = lower($2)
`

	user, err := scanUser(r.db.QueryRowContext(ctx, query, sourceID, username))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by source and username: %w", err)
	}

	return user, nil
}

func (r *UserRepository) FindBySubject(ctx context.Context, subject string) (*domain.User, error) {
	const query = `
SELECT id, subject, source_id, username, display_name, email, password_hash, is_active, last_login_at, created_at, updated_at
FROM auth.users
WHERE subject = $1
`

	user, err := scanUser(r.db.QueryRowContext(ctx, query, subject))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by subject: %w", err)
	}

	return user, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	const query = `
SELECT id, subject, source_id, username, display_name, email, password_hash, is_active, last_login_at, created_at, updated_at
FROM auth.users
WHERE id = $1
`

	user, err := scanUser(r.db.QueryRowContext(ctx, query, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	return user, nil
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	if user == nil {
		return nil, errors.New("user is nil")
	}

	const query = `
INSERT INTO auth.users (
	subject,
	source_id,
	username,
	display_name,
	email,
	password_hash,
	is_active,
	last_login_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, created_at, updated_at
`

	var (
		id        string
		createdAt time.Time
		updatedAt time.Time
	)

	if err := r.db.QueryRowContext(
		ctx,
		query,
		user.Subject,
		user.SourceID,
		user.Username,
		stringToNull(user.DisplayName),
		stringToNull(user.Email),
		stringToNull(user.PasswordHash),
		user.IsActive,
		user.LastLoginAt,
	).Scan(&id, &createdAt, &updatedAt); err != nil {
		if isUniqueViolation(err) {
			return nil, domain.ErrUserExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	created := *user
	created.ID = id
	created.CreatedAt = createdAt
	created.UpdatedAt = updatedAt

	return &created, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id string, at time.Time) error {
	const query = `
UPDATE auth.users
SET last_login_at = $2, updated_at = $2
WHERE id = $1
`

	result, err := r.db.ExecContext(ctx, query, id, at)
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read affected rows for update last login: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) Update(
	ctx context.Context,
	id string,
	displayName *string,
	email *string,
	passwordHash *string,
	isActive *bool,
) (*domain.User, error) {
	const query = `
UPDATE auth.users
SET display_name = COALESCE($2, display_name),
    email = COALESCE($3, email),
    password_hash = COALESCE($4, password_hash),
    is_active = COALESCE($5, is_active),
    updated_at = now()
WHERE id = $1
`

	result, err := r.db.ExecContext(
		ctx,
		query,
		id,
		nullableString(displayName),
		nullableString(email),
		nullableString(passwordHash),
		isActive,
	)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read affected rows for user update: %w", err)
	}
	if rowsAffected == 0 {
		return nil, domain.ErrUserNotFound
	}

	updatedUser, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("load updated user: %w", err)
	}
	if updatedUser == nil {
		return nil, domain.ErrUserNotFound
	}

	return updatedUser, nil
}

func (r *UserRepository) List(
	ctx context.Context,
	filter UserListFilter,
) ([]*domain.User, int, error) {
	whereParts := make([]string, 0, 3)
	args := make([]any, 0, 8)
	countArgs := make([]any, 0, 6)

	if value := strings.TrimSpace(filter.Query); value != "" {
		whereParts = append(
			whereParts,
			fmt.Sprintf(
				"(username ILIKE $%d OR email ILIKE $%d)",
				len(args)+1,
				len(args)+2,
			),
		)
		likeValue := "%" + value + "%"
		args = append(args, likeValue, likeValue)
		countArgs = append(countArgs, likeValue, likeValue)
	}
	if filter.SourceID != nil {
		whereParts = append(whereParts, fmt.Sprintf("source_id = $%d", len(args)+1))
		args = append(args, *filter.SourceID)
		countArgs = append(countArgs, *filter.SourceID)
	}
	if filter.IsActive != nil {
		whereParts = append(whereParts, fmt.Sprintf("is_active = $%d", len(args)+1))
		args = append(args, *filter.IsActive)
		countArgs = append(countArgs, *filter.IsActive)
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = " WHERE " + strings.Join(whereParts, " AND ")
	}

	countQuery := "SELECT count(*) FROM auth.users" + whereClause
	total := 0
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}

	listQuery := `
SELECT id, subject, source_id, username, display_name, email, password_hash, is_active, last_login_at, created_at, updated_at
FROM auth.users` + whereClause + `
ORDER BY created_at DESC
OFFSET $` + fmt.Sprintf("%d", len(args)+1) + ` LIMIT $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, filter.Offset, filter.Limit)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		user, scanErr := scanUser(rows)
		if scanErr != nil {
			return nil, 0, fmt.Errorf("scan user from list: %w", scanErr)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate users list: %w", err)
	}

	return users, total, nil
}

func (r *UserRepository) CountActiveAdminsExcluding(
	ctx context.Context,
	excludeUserID string,
) (int, error) {
	const query = `
SELECT count(*)
FROM auth.users u
JOIN auth.user_roles ur ON ur.user_id = u.id
WHERE ur.role_code = 'admin'
  AND u.is_active = true
  AND ($1 = '' OR u.id::text <> $1)
`

	total := 0
	if err := r.db.QueryRowContext(ctx, query, strings.TrimSpace(excludeUserID)).Scan(&total); err != nil {
		return 0, fmt.Errorf("count active admins: %w", err)
	}

	return total, nil
}

func (r *UserRepository) ListRolesCatalog(ctx context.Context) ([]domain.Role, error) {
	const query = `
SELECT code, name, description
FROM auth.roles
ORDER BY code
`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list roles catalog: %w", err)
	}
	defer rows.Close()

	roles := make([]domain.Role, 0)
	for rows.Next() {
		var role domain.Role
		var description sql.NullString
		if err := rows.Scan(&role.Code, &role.Name, &description); err != nil {
			return nil, fmt.Errorf("scan role catalog row: %w", err)
		}
		if description.Valid {
			role.Description = description.String
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate roles catalog: %w", err)
	}

	return roles, nil
}

func (r *UserRepository) GetRoles(ctx context.Context, userID string) ([]string, error) {
	const query = `
SELECT role_code
FROM auth.user_roles
WHERE user_id = $1
ORDER BY role_code
`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}
	defer rows.Close()

	roles := make([]string, 0)
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, fmt.Errorf("scan user role: %w", err)
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user roles: %w", err)
	}

	return roles, nil
}

func (r *UserRepository) SetRoles(
	ctx context.Context,
	userID string,
	roles []string,
	grantedBy *string,
) error {
	uniqueRoles := make([]string, 0, len(roles))
	seen := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		if !domain.IsAllowedRole(role) {
			return fmt.Errorf("role %q is not allowed", role)
		}

		if _, ok := seen[role]; ok {
			continue
		}
		seen[role] = struct{}{}
		uniqueRoles = append(uniqueRoles, role)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin set roles transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.ExecContext(
		ctx,
		`DELETE FROM auth.user_roles WHERE user_id = $1`,
		userID,
	); err != nil {
		return fmt.Errorf("delete existing user roles: %w", err)
	}

	if len(uniqueRoles) > 0 {
		const insertQuery = `
INSERT INTO auth.user_roles (user_id, role_code, granted_by)
VALUES ($1, $2, $3)
`

		for _, role := range uniqueRoles {
			if _, err := tx.ExecContext(ctx, insertQuery, userID, role, grantedBy); err != nil {
				return fmt.Errorf("insert user role %q: %w", role, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit set roles transaction: %w", err)
	}

	return nil
}

func scanUser(scanner interface{ Scan(dest ...any) error }) (*domain.User, error) {
	var (
		user         domain.User
		displayName  sql.NullString
		email        sql.NullString
		passwordHash sql.NullString
		lastLoginAt  sql.NullTime
	)

	if err := scanner.Scan(
		&user.ID,
		&user.Subject,
		&user.SourceID,
		&user.Username,
		&displayName,
		&email,
		&passwordHash,
		&user.IsActive,
		&lastLoginAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}

	if displayName.Valid {
		user.DisplayName = displayName.String
	}
	if email.Valid {
		user.Email = email.String
	}
	if passwordHash.Valid {
		user.PasswordHash = passwordHash.String
	}
	if lastLoginAt.Valid {
		loginAt := lastLoginAt.Time
		user.LastLoginAt = &loginAt
	}

	return &user, nil
}

func stringToNull(value string) any {
	if value == "" {
		return nil
	}

	return value
}

func nullableString(value *string) any {
	if value == nil {
		return nil
	}

	text := strings.TrimSpace(*value)
	if text == "" {
		return ""
	}

	return text
}
