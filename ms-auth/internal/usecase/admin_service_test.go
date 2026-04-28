package usecase

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/EthernalFox/Controlitix/ms-auth/internal/domain"
)

type adminUsersMock struct {
	findByID                  func(ctx context.Context, id string) (*domain.User, error)
	findBySubject             func(ctx context.Context, subject string) (*domain.User, error)
	create                    func(ctx context.Context, user *domain.User) (*domain.User, error)
	update                    func(ctx context.Context, id string, displayName *string, email *string, passwordHash *string, isActive *bool) (*domain.User, error)
	getRoles                  func(ctx context.Context, userID string) ([]string, error)
	setRoles                  func(ctx context.Context, userID string, roles []string, grantedBy *string) error
	countActiveAdminsExcluding func(ctx context.Context, excludeUserID string) (int, error)
}

func (m *adminUsersMock) FindByID(ctx context.Context, id string) (*domain.User, error) {
	return m.findByID(ctx, id)
}

func (m *adminUsersMock) FindBySubject(ctx context.Context, subject string) (*domain.User, error) {
	return m.findBySubject(ctx, subject)
}

func (m *adminUsersMock) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	return m.create(ctx, user)
}

func (m *adminUsersMock) Update(
	ctx context.Context,
	id string,
	displayName *string,
	email *string,
	passwordHash *string,
	isActive *bool,
) (*domain.User, error) {
	return m.update(ctx, id, displayName, email, passwordHash, isActive)
}

func (m *adminUsersMock) GetRoles(ctx context.Context, userID string) ([]string, error) {
	return m.getRoles(ctx, userID)
}

func (m *adminUsersMock) SetRoles(
	ctx context.Context,
	userID string,
	roles []string,
	grantedBy *string,
) error {
	return m.setRoles(ctx, userID, roles, grantedBy)
}

func (m *adminUsersMock) CountActiveAdminsExcluding(
	ctx context.Context,
	excludeUserID string,
) (int, error) {
	return m.countActiveAdminsExcluding(ctx, excludeUserID)
}

type adminServicesMock struct {
	findByID         func(ctx context.Context, id string) (*domain.ServiceAccount, error)
	create           func(ctx context.Context, account *domain.ServiceAccount) (*domain.ServiceAccount, error)
	update           func(ctx context.Context, account *domain.ServiceAccount) error
	updateSecretHash func(ctx context.Context, id string, hash string) error
	delete           func(ctx context.Context, id string) error
}

func (m *adminServicesMock) FindByID(ctx context.Context, id string) (*domain.ServiceAccount, error) {
	return m.findByID(ctx, id)
}

func (m *adminServicesMock) Create(
	ctx context.Context,
	account *domain.ServiceAccount,
) (*domain.ServiceAccount, error) {
	return m.create(ctx, account)
}

func (m *adminServicesMock) Update(ctx context.Context, account *domain.ServiceAccount) error {
	return m.update(ctx, account)
}

func (m *adminServicesMock) UpdateSecretHash(
	ctx context.Context,
	id string,
	hash string,
) error {
	return m.updateSecretHash(ctx, id, hash)
}

func (m *adminServicesMock) Delete(ctx context.Context, id string) error {
	return m.delete(ctx, id)
}

type adminSourcesMock struct {
	findByID func(ctx context.Context, id int) (*domain.IdentitySource, error)
}

func (m *adminSourcesMock) FindByID(ctx context.Context, id int) (*domain.IdentitySource, error) {
	return m.findByID(ctx, id)
}

type adminRefreshMock struct {
	revokeAllForUser func(ctx context.Context, userID string, at time.Time, reason string) error
}

func (m *adminRefreshMock) RevokeAllForUser(
	ctx context.Context,
	userID string,
	at time.Time,
	reason string,
) error {
	return m.revokeAllForUser(ctx, userID, at, reason)
}

type adminAuditMock struct {
	events []domain.AuditEvent
}

func (m *adminAuditMock) Record(_ context.Context, event domain.AuditEvent) error {
	m.events = append(m.events, event)
	return nil
}

func TestAdminServiceCreateUser(t *testing.T) {
	t.Parallel()

	hasher := testHasher(t)
	audit := &adminAuditMock{}
	setRolesCalled := false
	var createdPasswordHash string

	service := &AdminService{
		users: &adminUsersMock{
			findByID: func(ctx context.Context, id string) (*domain.User, error) { return nil, errors.New("unexpected findByID call") },
			findBySubject: func(ctx context.Context, subject string) (*domain.User, error) {
				return nil, errors.New("unexpected findBySubject call")
			},
			create: func(ctx context.Context, user *domain.User) (*domain.User, error) {
				createdPasswordHash = user.PasswordHash
				return &domain.User{
					ID:          "user-id",
					Subject:     user.Subject,
					SourceID:    user.SourceID,
					Username:    user.Username,
					DisplayName: user.DisplayName,
					Email:       user.Email,
					IsActive:    true,
				}, nil
			},
			update: func(ctx context.Context, id string, displayName *string, email *string, passwordHash *string, isActive *bool) (*domain.User, error) {
				return nil, errors.New("unexpected update call")
			},
			getRoles: func(ctx context.Context, userID string) ([]string, error) {
				return nil, errors.New("unexpected getRoles call")
			},
			setRoles: func(ctx context.Context, userID string, roles []string, grantedBy *string) error {
				setRolesCalled = true
				if userID != "user-id" {
					t.Fatalf("unexpected user id: %s", userID)
				}
				if len(roles) != 1 || roles[0] != "engineer" {
					t.Fatalf("unexpected roles: %#v", roles)
				}
				return nil
			},
			countActiveAdminsExcluding: func(ctx context.Context, excludeUserID string) (int, error) {
				return 1, nil
			},
		},
		services: &adminServicesMock{},
		sources: &adminSourcesMock{
			findByID: func(ctx context.Context, id int) (*domain.IdentitySource, error) {
				return &domain.IdentitySource{
					ID:        id,
					Type:      "local",
					Name:      "Local",
					IsEnabled: true,
				}, nil
			},
		},
		audit: audit,
		refresh: &adminRefreshMock{
			revokeAllForUser: func(ctx context.Context, userID string, at time.Time, reason string) error {
				return errors.New("unexpected revoke call")
			},
		},
		hasher: hasher,
		clock:  time.Now,
		rand:   bytes.NewReader([]byte("0123456789abcdef")),
	}

	createdUser, err := service.CreateUser(context.Background(), "local:admin", CreateUserInput{
		Username:    "ivan.petrov",
		DisplayName: "Ivan Petrov",
		Email:       "ivan@example.local",
		Password:    "StrongPass1!",
		Roles:       []string{"engineer"},
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if createdUser.ID != "user-id" {
		t.Fatalf("unexpected created user id: %s", createdUser.ID)
	}
	if !setRolesCalled {
		t.Fatal("expected set roles call")
	}
	if createdPasswordHash == "" {
		t.Fatal("expected created password hash")
	}
	ok, verifyErr := hasher.Verify("StrongPass1!", createdPasswordHash)
	if verifyErr != nil {
		t.Fatalf("verify created password hash: %v", verifyErr)
	}
	if !ok {
		t.Fatal("expected password hash to match password")
	}
	if len(audit.events) != 1 || audit.events[0].Action != domain.AuditAdminUserCreate {
		t.Fatalf("unexpected audit events: %#v", audit.events)
	}
}

func TestAdminServiceCreateUserValidation(t *testing.T) {
	t.Parallel()

	service := &AdminService{}
	_, err := service.CreateUser(context.Background(), "local:admin", CreateUserInput{
		Username: "user",
		Password: "pass",
		Roles:    []string{"unknown"},
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation, got %v", err)
	}
}

func TestAdminServiceSetUserRolesLastAdmin(t *testing.T) {
	t.Parallel()

	service := &AdminService{
		users: &adminUsersMock{
			findByID: func(ctx context.Context, id string) (*domain.User, error) {
				return &domain.User{
					ID:       id,
					Subject:  "local:admin",
					IsActive: true,
				}, nil
			},
			findBySubject: func(ctx context.Context, subject string) (*domain.User, error) {
				return nil, nil
			},
			create: func(ctx context.Context, user *domain.User) (*domain.User, error) {
				return nil, errors.New("unexpected create call")
			},
			update: func(ctx context.Context, id string, displayName *string, email *string, passwordHash *string, isActive *bool) (*domain.User, error) {
				return nil, errors.New("unexpected update call")
			},
			getRoles: func(ctx context.Context, userID string) ([]string, error) {
				return []string{"admin"}, nil
			},
			setRoles: func(ctx context.Context, userID string, roles []string, grantedBy *string) error {
				return errors.New("unexpected set roles call")
			},
			countActiveAdminsExcluding: func(ctx context.Context, excludeUserID string) (int, error) {
				return 0, nil
			},
		},
	}

	err := service.SetUserRoles(context.Background(), "local:other-admin", "admin-id", []string{})
	if !errors.Is(err, domain.ErrLastAdmin) {
		t.Fatalf("expected ErrLastAdmin, got %v", err)
	}
}

func TestAdminServiceDeactivateUserLastAdmin(t *testing.T) {
	t.Parallel()

	service := &AdminService{
		users: &adminUsersMock{
			findByID: func(ctx context.Context, id string) (*domain.User, error) {
				return &domain.User{
					ID:       id,
					Subject:  "local:admin",
					IsActive: true,
				}, nil
			},
			findBySubject: func(ctx context.Context, subject string) (*domain.User, error) {
				return nil, nil
			},
			create: func(ctx context.Context, user *domain.User) (*domain.User, error) {
				return nil, errors.New("unexpected create call")
			},
			update: func(ctx context.Context, id string, displayName *string, email *string, passwordHash *string, isActive *bool) (*domain.User, error) {
				return nil, errors.New("unexpected update call")
			},
			getRoles: func(ctx context.Context, userID string) ([]string, error) {
				return []string{"admin"}, nil
			},
			setRoles: func(ctx context.Context, userID string, roles []string, grantedBy *string) error {
				return errors.New("unexpected set roles call")
			},
			countActiveAdminsExcluding: func(ctx context.Context, excludeUserID string) (int, error) {
				return 0, nil
			},
		},
	}

	err := service.DeactivateUser(context.Background(), "local:other-admin", "admin-id")
	if !errors.Is(err, domain.ErrLastAdmin) {
		t.Fatalf("expected ErrLastAdmin, got %v", err)
	}
}

func TestAdminServiceCreateAndRotateServiceSecret(t *testing.T) {
	t.Parallel()

	hasher := testHasher(t)
	audit := &adminAuditMock{}

	var firstHash string
	var secondHash string
	service := &AdminService{
		users: &adminUsersMock{},
		services: &adminServicesMock{
			findByID: func(ctx context.Context, id string) (*domain.ServiceAccount, error) {
				return &domain.ServiceAccount{
					ID:       id,
					ClientID: "ms-poll",
				}, nil
			},
			create: func(ctx context.Context, account *domain.ServiceAccount) (*domain.ServiceAccount, error) {
				firstHash = account.ClientSecretHash
				return &domain.ServiceAccount{
					ID:          "service-id",
					ClientID:    account.ClientID,
					DisplayName: account.DisplayName,
					Scopes:      account.Scopes,
					IsActive:    account.IsActive,
				}, nil
			},
			update: func(ctx context.Context, account *domain.ServiceAccount) error {
				return nil
			},
			updateSecretHash: func(ctx context.Context, id string, hash string) error {
				secondHash = hash
				return nil
			},
			delete: func(ctx context.Context, id string) error {
				return nil
			},
		},
		sources: &adminSourcesMock{},
		audit:   audit,
		refresh: &adminRefreshMock{},
		hasher:  hasher,
		clock:   time.Now,
		rand: bytes.NewReader(append(
			[]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
			[]byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")...,
		)),
	}

	account, secret, err := service.CreateServiceAccount(context.Background(), "local:admin", CreateServiceAccountInput{
		ClientID:    "ms-poll",
		DisplayName: "MS Poll",
		Scopes:      []string{"config.read", "tags.values.write"},
	})
	if err != nil {
		t.Fatalf("create service account: %v", err)
	}
	if account.ClientID != "ms-poll" {
		t.Fatalf("unexpected client id: %s", account.ClientID)
	}
	if secret == "" {
		t.Fatal("expected plaintext secret")
	}

	ok, verifyErr := hasher.Verify(secret, firstHash)
	if verifyErr != nil {
		t.Fatalf("verify first hash: %v", verifyErr)
	}
	if !ok {
		t.Fatal("expected first secret to match first hash")
	}

	rotatedSecret, err := service.RotateServiceAccountSecret(
		context.Background(),
		"local:admin",
		"service-id",
	)
	if err != nil {
		t.Fatalf("rotate service secret: %v", err)
	}

	if rotatedSecret == "" || rotatedSecret == secret {
		t.Fatal("expected non-empty rotated secret different from old one")
	}

	ok, verifyErr = hasher.Verify(rotatedSecret, secondHash)
	if verifyErr != nil {
		t.Fatalf("verify second hash: %v", verifyErr)
	}
	if !ok {
		t.Fatal("expected rotated secret to match rotated hash")
	}
	oldStillValid, verifyErr := hasher.Verify(secret, secondHash)
	if verifyErr != nil {
		t.Fatalf("verify old secret against rotated hash: %v", verifyErr)
	}
	if oldStillValid {
		t.Fatal("old secret must not match rotated hash")
	}

	if len(audit.events) < 2 {
		t.Fatalf("expected create and rotate audit events, got %#v", audit.events)
	}
}
