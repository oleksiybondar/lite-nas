package getfacl

import "testing"

func TestParseParsesHeadersAndOwnerGroup(t *testing.T) {
	t.Parallel()

	document, err := Parse(validInputFixture())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if document.FilePath != "/etc/lite-nas/certificates" {
		t.Fatalf("FilePath = %q", document.FilePath)
	}
	if document.Owner != "root" || document.Group != "lite-nas" {
		t.Fatalf("owner/group = %q/%q", document.Owner, document.Group)
	}
}

func TestParseParsesCoreAndMaskPermissions(t *testing.T) {
	t.Parallel()

	document, err := Parse(validInputFixture())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	assertUserPermissionRWX(t, document)
	assertMaskPermissionRX(t, document)
}

func TestParseParsesNamedUserAndGroupEntries(t *testing.T) {
	t.Parallel()

	document, err := Parse(validInputFixture())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	assertNamedEntryExists(t, document.NamedUsers, "testuser", "named user")
	assertNamedEntryExists(t, document.NamedGroups, "admins", "named group")
}

func TestParseParsesUserPermissionOnly(t *testing.T) {
	t.Parallel()

	document, err := Parse(validInputFixture())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	assertUserPermissionRWX(t, document)
}

func TestParseParsesMaskPermissionOnly(t *testing.T) {
	t.Parallel()

	document, err := Parse(validInputFixture())
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	assertMaskPermissionRX(t, document)
}

func TestParseReturnsErrorWhenOwnerGroupHeadersMissing(t *testing.T) {
	t.Parallel()

	_, err := Parse("user::rwx\nother::---\n")
	if err == nil {
		t.Fatalf("Parse() expected error for missing owner/group headers")
	}
}

func validInputFixture() string {
	return `# file: /etc/lite-nas/certificates
# owner: root
# group: lite-nas
user::rwx
user:testuser:r--
group::--x
group:admins:rwx
mask::r-x
other::--x
`
}

func assertUserPermissionRWX(t *testing.T, document Document) {
	t.Helper()
	if !document.User.Read || !document.User.Write || !document.User.Execute {
		t.Fatalf("user:: permission parse mismatch: %#v", document.User)
	}
}

func assertMaskPermissionRX(t *testing.T, document Document) {
	t.Helper()
	if document.Mask == nil || !document.Mask.Read || document.Mask.Write || !document.Mask.Execute {
		t.Fatalf("mask parse mismatch: %#v", document.Mask)
	}
}

func assertNamedEntryExists(t *testing.T, entries map[string]Permission, key string, label string) {
	t.Helper()
	if _, ok := entries[key]; !ok {
		t.Fatalf("%s entry missing", label)
	}
}
