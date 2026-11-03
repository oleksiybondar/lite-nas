package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"lite-nas/services/file-service/filesystem"
	authcontract "lite-nas/shared/contracts/auth"
	rbaccontract "lite-nas/shared/contracts/rbac"
	"lite-nas/shared/fileaccess"
)

// requestStub controls the Auth and RBAC calls made by one HTTP test.
type requestStub func(context.Context, string, any, any) error

func (stub requestStub) Request(ctx context.Context, subject string, request any, response any) error {
	return stub(ctx, subject, request, response)
}

// mustService builds a service with temporary storage and controlled RPC replies.
func mustService(t *testing.T, root string, requester requestStub) *fileService {
	t.Helper()
	policy, err := filesystem.NewPolicy([]string{root}, nil)
	if err != nil {
		t.Fatal(err)
	}
	auth, err := fileaccess.NewAuthorizer(requester)
	if err != nil {
		t.Fatal(err)
	}
	service, err := newFileService(serviceConfig{MetadataDir: t.TempDir(), ChunkSize: 1 << 20}, policy, auth)
	if err != nil {
		t.Fatal(err)
	}
	return service
}

// authorizedRequests grants Auth and RBAC while counting read decisions.
func authorizedRequests(reads *int, groups ...string) requestStub {
	return func(_ context.Context, rpcSubject string, _ any, response any) error {
		switch rpcSubject {
		case authcontract.ValidateAccessTokenRPCSubject:
			*response.(*authcontract.ValidateAccessTokenResponse) = authcontract.ValidateAccessTokenResponse{
				Valid: true, Status: authcontract.StatusAuthenticated, Username: "alice",
			}
		case rbaccontract.GetSubjectRolesRPCSubject:
			response.(*rbaccontract.GetSubjectRolesResponse).UID = "1001"
			response.(*rbaccontract.GetSubjectRolesResponse).Groups = groups
		case rbaccontract.CanReadPathRPCSubject:
			*reads++
			response.(*rbaccontract.DecisionResponse).Allowed = true
		default:
			return errors.New("unexpected RPC subject")
		}
		return nil
	}
}

// TestOperatorReadRoutesHideSystemPaths covers file-service-svc/FR-003 and FR-008.
func TestOperatorReadRoutesHideSystemPaths(t *testing.T) {
	service := mustService(t, t.TempDir(), authorizedRequests(new(int), "lite-nas-operator"))
	for _, test := range []struct {
		path       string
		wantStatus int
	}{
		{path: "/ls/", wantStatus: http.StatusOK},
		{path: "/ls/etc", wantStatus: http.StatusForbidden},
		{path: "/etc/passwd", wantStatus: http.StatusForbidden},
	} {
		request := httptest.NewRequest(http.MethodGet, test.path, nil)
		request.AddCookie(&http.Cookie{Name: "lite-nas-at", Value: "token"})
		response := httptest.NewRecorder()
		service.ServeHTTP(response, request)
		if response.Code != test.wantStatus {
			t.Errorf("GET %s = %d %q, want %d", test.path, response.Code, response.Body.String(), test.wantStatus)
		}
		if test.path == "/ls/" && strings.Contains(response.Body.String(), `"/etc"`) {
			t.Error("operator root listing exposed /etc")
		}
	}
}

// TestAdministratorRootListingKeepsSystemPaths covers file-service-svc/FR-008.
func TestAdministratorRootListingKeepsSystemPaths(t *testing.T) {
	service := mustService(t, t.TempDir(), authorizedRequests(new(int), "sudo"))
	request := httptest.NewRequest(http.MethodGet, "/ls/", nil)
	request.AddCookie(&http.Cookie{Name: "lite-nas-at", Value: "token"})
	response := httptest.NewRecorder()
	service.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"/etc"`) {
		t.Fatalf("administrator root listing = %d %q", response.Code, response.Body.String())
	}
}

// TestListingReturnsMetadataAndKeepsFlags covers file-service-svc/FR-008.
func TestListingReturnsMetadataAndKeepsFlags(t *testing.T) {
	root := t.TempDir()
	file := filepath.Join(root, "report.txt")
	if err := os.WriteFile(file, []byte("hello"), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "folder"), 0o750); err != nil {
		t.Fatal(err)
	}
	service := mustService(t, root, authorizedRequests(new(int), "sudo"))
	for _, test := range []struct {
		query       string
		wantEntries int
	}{
		{query: "", wantEntries: 2},
		{query: "?files-only", wantEntries: 1},
	} {
		request := httptest.NewRequest(http.MethodGet, "/ls"+root+test.query, nil)
		request.AddCookie(&http.Cookie{Name: "lite-nas-at", Value: "token"})
		response := httptest.NewRecorder()
		service.ServeHTTP(response, request)
		if response.Code != http.StatusOK {
			t.Fatalf("listing = %d %q", response.Code, response.Body.String())
		}
		var entries []listingEntry
		if err := json.Unmarshal(response.Body.Bytes(), &entries); err != nil {
			t.Fatal(err)
		}
		if len(entries) != test.wantEntries {
			t.Fatalf("listing entries = %d, want %d", len(entries), test.wantEntries)
		}
		for _, entry := range entries {
			if entry.Kind == "directory" && test.query != "" {
				t.Fatal("files-only returned a directory")
			}
			if entry.Path == file && (entry.Name != "report.txt" || entry.Size == nil || *entry.Size != 5 ||
				entry.Owner == nil || entry.Group == nil || entry.Permissions == nil || *entry.Permissions != "rw-r-----" ||
				entry.LastModified == nil) {
				t.Fatalf("file metadata = %+v", entry)
			}
		}
	}
}

// TestRecursiveListingOmitsConfiguredOSMetadata covers file-service-svc/FR-008.
func TestRecursiveListingOmitsConfiguredOSMetadata(t *testing.T) {
	root := t.TempDir()
	for _, directory := range []string{".Trashes", "pictures"} {
		if err := os.Mkdir(filepath.Join(root, directory), 0o750); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"Thumbs.DB", "desktop.INI", ".DS_Store", "._photo.jpg", ".directory", ".hidden", ".keep", "notes.txt", ".Trashes/deleted.txt", "pictures/photo.jpg"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("content"), 0o640); err != nil {
			t.Fatal(err)
		}
	}
	service := mustService(t, root, authorizedRequests(new(int), "sudo"))
	service.cfg.ListingIgnoredNames = mustLoadPackagedConfig(t).ListingIgnoredNames
	request := httptest.NewRequest(http.MethodGet, "/ls"+root+"?recursive&files-only", nil)
	request.AddCookie(&http.Cookie{Name: "lite-nas-at", Value: "token"})
	response := httptest.NewRecorder()
	service.ServeHTTP(response, request)
	var entries []listingEntry
	if err := json.Unmarshal(response.Body.Bytes(), &entries); err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name)
	}
	sort.Strings(names)
	if !slices.Equal(names, []string{".keep", "notes.txt", "photo.jpg"}) {
		t.Fatalf("listing names = %v", names)
	}
}

// TestListingFilterDoesNotDenyDirectReads covers the listing-only filter contract.
func TestListingFilterDoesNotDenyDirectReads(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "Thumbs.db"), []byte("content"), 0o640); err != nil {
		t.Fatal(err)
	}
	service := mustService(t, root, authorizedRequests(new(int), "sudo"))
	service.cfg.ListingIgnoredNames = mustLoadPackagedConfig(t).ListingIgnoredNames
	service.readerCommand = openWorkflowFile
	request := httptest.NewRequest(http.MethodGet, filepath.Join(root, "Thumbs.db"), nil)
	request.AddCookie(&http.Cookie{Name: "lite-nas-at", Value: "token"})
	response := httptest.NewRecorder()
	service.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("direct read status = %d", response.Code)
	}
}

// TestMissingTokenNeverReadsFile covers file-service-svc/FR-001 and SR-001.
func TestMissingTokenNeverReadsFile(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "private.txt")
	if err := os.WriteFile(path, []byte("secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	service := mustService(t, root, func(_ context.Context, _ string, _ any, _ any) error {
		t.Fatal("missing token reached RPC")
		return nil
	})
	request := httptest.NewRequest(http.MethodGet, "/"+strings.TrimPrefix(path, "/"), nil)
	response := httptest.NewRecorder()
	service.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized || strings.Contains(response.Body.String(), "secret") {
		t.Fatalf("response = %d %q", response.Code, response.Body.String())
	}
}

// TestMutationPolicyDenialsAreForbidden covers file-service-svc/FR-004.
func TestMutationPolicyDenialsAreForbidden(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
	}{
		{name: "protected path", err: filesystem.ErrProtectedPath},
		{name: "outside write roots", err: filesystem.ErrOutsideWriteRoots},
		{name: "filesystem permission", err: os.ErrPermission},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			decisionError(response, test.err)
			if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), `"code":"forbidden"`) {
				t.Fatalf("policy denial = %d %q", response.Code, response.Body.String())
			}
		})
	}
}

// TestAuthenticationFailuresKeepDistinctStatuses covers file-service-svc/FR-001.
func TestAuthenticationFailuresKeepDistinctStatuses(t *testing.T) {
	for _, test := range []struct {
		name       string
		failedRPC  string
		invalidUID bool
		wantStatus int
		wantCode   string
	}{
		{
			name: "Auth RPC unavailable", failedRPC: authcontract.ValidateAccessTokenRPCSubject,
			wantStatus: http.StatusServiceUnavailable, wantCode: "authentication_unavailable",
		},
		{
			name: "RBAC RPC unavailable", failedRPC: rbaccontract.GetSubjectRolesRPCSubject,
			wantStatus: http.StatusServiceUnavailable, wantCode: "authentication_unavailable",
		},
		{
			name: "RBAC UID unavailable", invalidUID: true,
			wantStatus: http.StatusServiceUnavailable, wantCode: "authentication_unavailable",
		},
		{name: "invalid access token", wantStatus: http.StatusUnauthorized, wantCode: "unauthenticated"},
	} {
		t.Run(test.name, func(t *testing.T) {
			service := mustService(t, t.TempDir(), func(_ context.Context, subject string, _ any, response any) error {
				if subject == test.failedRPC {
					return errors.New("messaging unavailable")
				}
				switch subject {
				case authcontract.ValidateAccessTokenRPCSubject:
					if test.wantStatus == http.StatusUnauthorized {
						return nil
					}
					*response.(*authcontract.ValidateAccessTokenResponse) = authcontract.ValidateAccessTokenResponse{
						Valid: true, Status: authcontract.StatusAuthenticated, Username: "alice",
					}
				case rbaccontract.GetSubjectRolesRPCSubject:
					if !test.invalidUID {
						response.(*rbaccontract.GetSubjectRolesResponse).UID = "1001"
					}
				default:
					t.Fatalf("unexpected RPC subject %q", subject)
				}
				return nil
			})
			request := httptest.NewRequest(http.MethodGet, "/private.txt", nil)
			request.AddCookie(&http.Cookie{Name: "lite-nas-at", Value: "token"})
			response := httptest.NewRecorder()
			service.ServeHTTP(response, request)
			if response.Code != test.wantStatus ||
				!strings.Contains(response.Body.String(), `"`+test.wantCode+`"`) ||
				strings.Contains(response.Body.String(), "messaging unavailable") {
				t.Fatalf("response = %d %q", response.Code, response.Body.String())
			}
		})
	}
}

// TestAccessCookieAuthenticatesRead covers file-service-svc/FR-001.
func TestAccessCookieAuthenticatesRead(t *testing.T) {
	service := mustService(t, t.TempDir(), authorizedRequests(new(int)))
	request := httptest.NewRequest(http.MethodGet, "/proprietary/upload-window", nil)
	request.AddCookie(&http.Cookie{Name: "lite-nas-at", Value: "token"})
	response := httptest.NewRecorder()
	service.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("response = %d %q", response.Code, response.Body.String())
	}
}

// TestCaptionGrantReadsOnlyListedSource covers the File Service-only grant route.
func TestCaptionGrantReadsOnlyListedSource(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "lizard.jpg")
	if err := os.WriteFile(path, []byte("image"), 0o600); err != nil {
		t.Fatal(err)
	}
	service := mustService(t, root, requestStub(func(_ context.Context, subject string, request any, response any) error {
		switch subject {
		case authcontract.CaptionGrantValidateRPCSubject:
			input := request.(authcontract.CaptionGrantValidateRequest)
			*response.(*authcontract.CaptionGrantValidateResponse) = authcontract.CaptionGrantValidateResponse{Valid: input.Path == path, Username: "alice"}
		case rbaccontract.GetSubjectRolesRPCSubject:
			*response.(*rbaccontract.GetSubjectRolesResponse) = rbaccontract.GetSubjectRolesResponse{UID: "1001"}
		case rbaccontract.CanReadPathRPCSubject:
			response.(*rbaccontract.DecisionResponse).Allowed = true
		default:
			return errors.New("unexpected RPC")
		}
		return nil
	}))
	service.readerCommand = openWorkflowFile
	other := filepath.Join(root, "other.jpg")
	if err := os.WriteFile(other, []byte("other"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		path   string
		status int
	}{{path, http.StatusOK}, {other, http.StatusUnauthorized}} {
		request := httptest.NewRequest(http.MethodGet, test.path, nil)
		request.Header.Set("Authorization", "CaptionGrant token")
		response := httptest.NewRecorder()
		service.ServeHTTP(response, request)
		if response.Code != test.status {
			t.Fatalf("caption read %s status = %d", test.path, response.Code)
		}
	}
}

// TestCaptionGrantRejectsMutation keeps the credential read-only.
func TestCaptionGrantRejectsMutation(t *testing.T) {
	service := mustService(t, t.TempDir(), authorizedRequests(new(int)))
	request := httptest.NewRequest(http.MethodDelete, "/photos/lizard.jpg", nil)
	request.Header.Set("Authorization", "CaptionGrant token")
	response := httptest.NewRecorder()
	service.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("caption mutation status = %d", response.Code)
	}
}

// TestCaptionGrantRejectsOtherRoutes keeps the credential off File Service APIs.
func TestCaptionGrantRejectsOtherRoutes(t *testing.T) {
	service := mustService(t, t.TempDir(), authorizedRequests(new(int)))
	request := httptest.NewRequest(http.MethodGet, "/profile/ui", nil)
	request.Header.Set("Authorization", "CaptionGrant token")
	response := httptest.NewRecorder()
	service.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("caption API status = %d", response.Code)
	}
}

// TestHeadUsesFreshRBACDecision covers file-service-svc/FR-003 and FR-006.
func TestHeadUsesFreshRBACDecision(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "photo.jpg")
	if err := os.WriteFile(path, []byte("photo"), 0o600); err != nil {
		t.Fatal(err)
	}
	reads := 0
	service := mustService(t, root, authorizedRequests(&reads))
	for _, mount := range []string{"/", "/download/"} {
		request := httptest.NewRequest(http.MethodHead, mount+strings.TrimPrefix(path, "/"), nil)
		request.Header.Set("Authorization", "Bearer token")
		response := httptest.NewRecorder()
		service.ServeHTTP(response, request)
		if response.Code != http.StatusOK || response.Body.Len() != 0 || response.Header().Get("ETag") == "" {
			t.Fatalf("%s response = %d, headers = %v", mount, response.Code, response.Header())
		}
	}
	if reads != 2 {
		t.Fatalf("RBAC read decisions = %d, want 2", reads)
	}
}

// TestRangeParserRejectsMultipleRanges covers file-service-svc/FR-006.
func TestRangeParserRejectsMultipleRanges(t *testing.T) {
	if _, _, _, err := parseRange("bytes=0-1,3-4", 10); err == nil {
		t.Fatal("multipart range was accepted")
	}
	start, end, partial, err := parseRange("bytes=-3", 10)
	if err != nil || !partial || start != 7 || end != 9 {
		t.Fatalf("suffix range = %d-%d, partial %t, error %v", start, end, partial, err)
	}
}

// TestCreateMetadataValidation covers file-service-svc/IR-001 input validation.
func TestCreateMetadataValidation(t *testing.T) {
	valid := `{"destination":"/home/alice/photo.jpg","filename":"photo.jpg","size":0,"sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`
	for _, test := range []struct {
		name string
		body string
		want bool
	}{
		{name: "valid", body: valid, want: true},
		{name: "missing filename", body: `{"destination":"/home/alice/photo.jpg","size":0,"sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`},
		{name: "negative size", body: `{"destination":"/home/alice/photo.jpg","filename":"photo.jpg","size":-1,"sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`},
		{name: "unknown field", body: strings.TrimSuffix(valid, "}") + `,"mode":"unsafe"}`},
		{name: "invalid private", body: strings.TrimSuffix(valid, "}") + `,"private":"yes"}`},
	} {
		t.Run(test.name, func(t *testing.T) {
			var input createInput
			if got := validateCreateDTO([]byte(test.body), &input); got != test.want {
				t.Fatalf("validation = %t, want %t", got, test.want)
			}
		})
	}
}
