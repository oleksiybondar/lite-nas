package identity

import "testing"

func TestParseParsesDefaultIDOutput(t *testing.T) {
	t.Parallel()

	idOutput := "uid=1002(testuser) gid=1002(testgroup) groups=1002(testgroup),27(sudo),4(adm)"

	identity, err := Parse(idOutput)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	assertIdentityIDs(t, identity, "1002", "1002")
}

func TestParseParsesPrimaryNames(t *testing.T) {
	t.Parallel()

	identity, err := Parse("uid=1002(testuser) gid=1002(testgroup) groups=1002(testgroup),27(sudo),4(adm)")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if identity.Username != "testuser" || identity.PrimaryGroup != "testgroup" {
		t.Fatalf("name parse mismatch: %#v", identity)
	}
}

func TestParseParsesGroups(t *testing.T) {
	t.Parallel()

	identity, err := Parse("uid=1002(testuser) gid=1002(testgroup) groups=1002(testgroup),27(sudo),4(adm)")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(identity.Groups) != 3 || identity.Groups[0] != "testgroup" || identity.Groups[1] != "sudo" || identity.Groups[2] != "adm" {
		t.Fatalf("groups parse mismatch: %#v", identity.Groups)
	}
}

func TestParseReturnsErrorOnMissingFields(t *testing.T) {
	t.Parallel()

	_, err := Parse("uid=1002(testuser)")
	if err == nil {
		t.Fatalf("Parse() expected error for missing fields")
	}
}

func assertIdentityIDs(t *testing.T, identity Identity, uid string, gid string) {
	t.Helper()
	if identity.UID != uid || identity.GID != gid {
		t.Fatalf("uid/gid parse mismatch: %#v", identity)
	}
}
