package authtoken

import (
	"encoding/json"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func TestAccessClaimsMarshalIncludesIdentityAndAuthorizationFields(t *testing.T) {
	t.Parallel()

	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:  "lite-nas-auth",
			Subject: "1000",
		},
		Login:  "alice",
		Scopes: []string{"monitoring.read"},
		Roles:  []string{"operator"},
	}

	data, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var got map[string]any
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	assertJSONField(t, got, "iss", "lite-nas-auth")
	assertJSONField(t, got, "sub", "1000")
	assertJSONField(t, got, "login", "alice")
	assertJSONArrayField(t, got, "scopes", "monitoring.read")
	assertJSONArrayField(t, got, "roles", "operator")
}

func TestAccessClaimsMarshalIncludesEmptyAuthorizationArrays(t *testing.T) {
	t.Parallel()

	got := mustMarshalClaimsMap(t, AccessClaims{Login: "alice", Scopes: []string{}, Roles: []string{}})
	assertEmptyJSONArrayField(t, got, "scopes")
	assertEmptyJSONArrayField(t, got, "roles")
}

func assertJSONField(t *testing.T, fields map[string]any, name string, want string) {
	t.Helper()

	got, ok := fields[name].(string)
	if !ok || got != want {
		t.Fatalf("%s = %#v, want %q", name, fields[name], want)
	}
}

func assertJSONArrayField(t *testing.T, fields map[string]any, name string, want string) {
	t.Helper()

	values, ok := fields[name].([]any)
	if !ok || len(values) != 1 || values[0] != want {
		t.Fatalf("%s = %#v, want [%q]", name, fields[name], want)
	}
}

func assertEmptyJSONArrayField(t *testing.T, fields map[string]any, name string) {
	t.Helper()

	values, ok := fields[name].([]any)
	if !ok || len(values) != 0 {
		t.Fatalf("%s = %#v, want []", name, fields[name])
	}
}

func mustMarshalClaimsMap(t *testing.T, claims AccessClaims) map[string]any {
	t.Helper()

	data, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var got map[string]any
	if err = json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	return got
}
