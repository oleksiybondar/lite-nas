package sessions

import (
	"errors"
	"testing"
	"time"

	"lite-nas/shared/testutil/testcasetest"
)

func TestStoreCreateReturnsRefreshTokenAndRecord(t *testing.T) {
	t.Parallel()

	now := time.Unix(100, 0)
	store := NewStore(func() time.Time { return now }, StoreOptions{})

	token, record, err := store.Create(createInputFixture())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	testCases := []testcasetest.FieldCase[RefreshRecord]{
		{Name: "token hash", Got: func(record RefreshRecord) any { return record.TokenHash != "" }, Want: true},
		{Name: "expires at", Got: func(record RefreshRecord) any { return record.ExpiresAt }, Want: now.UTC().Add(30 * 24 * time.Hour)},
		{Name: "session id", Got: func(record RefreshRecord) any { return record.SessionID }, Want: "session-id"},
		{Name: "subject", Got: func(record RefreshRecord) any { return record.Subject }, Want: "1000"},
		{Name: "login", Got: func(record RefreshRecord) any { return record.Login }, Want: "alice"},
		{Name: "client ip", Got: func(record RefreshRecord) any { return record.ClientIP }, Want: "192.168.1.10"},
		{Name: "user agent", Got: func(record RefreshRecord) any { return record.UserAgent }, Want: "browser"},
	}

	if token.Value == "" {
		t.Fatal("refresh token value is empty")
	}
	if token.ExpiresAt != record.ExpiresAt {
		t.Fatalf("token ExpiresAt = %v, want %v", token.ExpiresAt, record.ExpiresAt)
	}
	testcasetest.RunFieldCases(t, func(*testing.T) RefreshRecord { return record }, testCases)
}

func TestStoreRotateReplacesTokenAndPreservesExpiry(t *testing.T) {
	t.Parallel()

	currentTime := time.Unix(100, 0)
	store := NewStore(func() time.Time { return currentTime }, StoreOptions{})
	context := refreshContextFixture()
	token, created := mustCreate(t, store)

	currentTime = currentTime.Add(time.Hour)
	rotatedToken, rotatedRecord, err := store.Rotate(token.Value, context)
	if err != nil {
		t.Fatalf("Rotate() error = %v", err)
	}

	if rotatedToken.Value == token.Value {
		t.Fatal("rotated token reused the previous token value")
	}
	if rotatedRecord.TokenHash == created.TokenHash {
		t.Fatal("rotated record reused the previous token hash")
	}
	if rotatedRecord.ExpiresAt != created.ExpiresAt {
		t.Fatalf("rotated ExpiresAt = %v, want %v", rotatedRecord.ExpiresAt, created.ExpiresAt)
	}
	if _, _, err = store.Rotate(token.Value, context); !errors.Is(err, ErrUnknownRefreshToken) {
		t.Fatalf("Rotate(old token) error = %v, want %v", err, ErrUnknownRefreshToken)
	}
}

func TestStoreRotateRejectsExpiredTokenAndRemovesRecord(t *testing.T) {
	t.Parallel()

	currentTime := time.Unix(100, 0)
	store := NewStore(func() time.Time { return currentTime }, StoreOptions{})
	token, _ := mustCreate(t, store)

	currentTime = currentTime.Add(31 * 24 * time.Hour)
	if _, _, err := store.Rotate(token.Value, refreshContextFixture()); !errors.Is(err, ErrExpiredRefreshToken) {
		t.Fatalf("Rotate() error = %v, want %v", err, ErrExpiredRefreshToken)
	}
	if got := store.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
}

func TestStoreRotateRejectsUserAgentMismatchAndRemovesRecord(t *testing.T) {
	t.Parallel()

	store := NewStore(time.Now, StoreOptions{})
	token, _ := mustCreate(t, store)
	context := refreshContextFixture()
	context.UserAgent = "different-browser"

	if _, _, err := store.Rotate(token.Value, context); !errors.Is(err, ErrRefreshTokenContextMismatch) {
		t.Fatalf("Rotate() error = %v, want %v", err, ErrRefreshTokenContextMismatch)
	}
	if got := store.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
}

func TestStoreRotateCanEnforceClientIP(t *testing.T) {
	t.Parallel()

	store := NewStore(time.Now, StoreOptions{EnforceClientIP: true})
	token, _ := mustCreate(t, store)
	context := refreshContextFixture()
	context.ClientIP = "192.168.1.11"

	if _, _, err := store.Rotate(token.Value, context); !errors.Is(err, ErrRefreshTokenContextMismatch) {
		t.Fatalf("Rotate() error = %v, want %v", err, ErrRefreshTokenContextMismatch)
	}
}

func TestStoreRotateUpdatesClientIPWhenNotEnforced(t *testing.T) {
	t.Parallel()

	store := NewStore(time.Now, StoreOptions{})
	token, _ := mustCreate(t, store)
	context := refreshContextFixture()
	context.ClientIP = "192.168.1.11"

	_, record, err := store.Rotate(token.Value, context)
	if err != nil {
		t.Fatalf("Rotate() error = %v", err)
	}
	if record.ClientIP != "192.168.1.11" {
		t.Fatalf("ClientIP = %q, want updated IP", record.ClientIP)
	}
}

func TestStoreRevokeDeletesToken(t *testing.T) {
	t.Parallel()

	store := NewStore(time.Now, StoreOptions{})
	token, _ := mustCreate(t, store)

	if err := store.Revoke(token.Value); err != nil {
		t.Fatalf("Revoke() error = %v", err)
	}
	if got := store.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
	if err := store.Revoke(token.Value); !errors.Is(err, ErrUnknownRefreshToken) {
		t.Fatalf("Revoke() error = %v, want %v", err, ErrUnknownRefreshToken)
	}
}

func TestStoreClearRemovesAllRecords(t *testing.T) {
	t.Parallel()

	store := NewStore(time.Now, StoreOptions{})
	mustCreate(t, store)
	mustCreate(t, store)

	store.Clear()
	if got := store.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
}

func TestStoreCleanupExpiredRemovesExpiredRecords(t *testing.T) {
	t.Parallel()

	currentTime := time.Unix(100, 0)
	store := NewStore(func() time.Time { return currentTime }, StoreOptions{})
	mustCreate(t, store)
	mustCreate(t, store)

	currentTime = currentTime.Add(31 * 24 * time.Hour)
	if removed := store.CleanupExpired(); removed != 2 {
		t.Fatalf("CleanupExpired() = %d, want 2", removed)
	}
	if got := store.Len(); got != 0 {
		t.Fatalf("Len() = %d, want 0", got)
	}
}

func TestStoreCreateRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	testCases := []CreateInput{
		{},
		{SessionID: "session-id", Subject: "1000", Login: "alice", TTL: time.Minute},
		{SessionID: "session-id", Subject: "1000", Login: "alice", Context: refreshContextFixture()},
	}

	for _, testCase := range testCases {
		store := NewStore(time.Now, StoreOptions{})
		if _, _, err := store.Create(testCase); !errors.Is(err, ErrInvalidRefreshSession) {
			t.Fatalf("Create() error = %v, want %v", err, ErrInvalidRefreshSession)
		}
	}
}

func mustCreate(t *testing.T, store *Store) (RefreshToken, RefreshRecord) {
	t.Helper()

	token, record, err := store.Create(createInputFixture())
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	return token, record
}

func createInputFixture() CreateInput {
	return CreateInput{
		SessionID: "session-id",
		Subject:   "1000",
		Login:     "alice",
		Context:   refreshContextFixture(),
		TTL:       30 * 24 * time.Hour,
	}
}

func refreshContextFixture() RefreshContext {
	return RefreshContext{
		ClientIP:  "192.168.1.10",
		UserAgent: "browser",
	}
}
