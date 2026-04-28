package sessions

import (
	"testing"
	"time"

	"lite-nas/shared/testutil/testcasetest"
)

func TestRefreshSessionKeepsOpaqueTokenHashAndLifecycleFields(t *testing.T) {
	t.Parallel()

	createdAt := time.Unix(100, 0)
	expiresAt := time.Unix(200, 0)
	rotatedAt := time.Unix(150, 0)
	revokedAt := time.Unix(175, 0)

	session := RefreshSession{
		ID:            "refresh-session-id",
		Subject:       "1000",
		Login:         "alice",
		TokenHash:     []byte("token-hash"),
		ExpiresAt:     expiresAt,
		CreatedAt:     createdAt,
		LastRotatedAt: rotatedAt,
		RevokedAt:     &revokedAt,
	}

	testCases := []testcasetest.FieldCase[RefreshSession]{
		{Name: "id", Got: func(session RefreshSession) any { return session.ID }, Want: "refresh-session-id"},
		{Name: "subject", Got: func(session RefreshSession) any { return session.Subject }, Want: "1000"},
		{Name: "login", Got: func(session RefreshSession) any { return session.Login }, Want: "alice"},
		{Name: "token hash", Got: func(session RefreshSession) any { return string(session.TokenHash) }, Want: "token-hash"},
		{Name: "expires at", Got: func(session RefreshSession) any { return session.ExpiresAt }, Want: expiresAt},
		{Name: "created at", Got: func(session RefreshSession) any { return session.CreatedAt }, Want: createdAt},
		{Name: "last rotated at", Got: func(session RefreshSession) any { return session.LastRotatedAt }, Want: rotatedAt},
		{Name: "revoked at", Got: func(session RefreshSession) any { return *session.RevokedAt }, Want: revokedAt},
	}

	testcasetest.RunFieldCases(t, func(*testing.T) RefreshSession { return session }, testCases)
}
