package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"lite-nas/shared/contracts/auth"
	"lite-nas/shared/contracts/rbac"
	"lite-nas/shared/fileaccess"
)

// TestAuthorizeReadUsesApplicationFolderUnlessGloballyEnforced covers file-service-svc/FR-003.
func TestAuthorizeReadUsesApplicationFolderUnlessGloballyEnforced(t *testing.T) {
	const folder = "/home/alice/gallery"
	for _, test := range []struct {
		name string
		app  string
		path string
		want string
	}{
		{name: "whitelisted app", app: "moments", path: folder + "/photo.jpg", want: folder},
		{name: "different app", app: "admin-panel", path: folder + "/photo.jpg", want: folder + "/photo.jpg"},
		{name: "globally enforced descendant", app: "moments", path: folder + "/private/photo.jpg", want: folder + "/private/photo.jpg"},
		{name: "prefix sibling", app: "moments", path: folder + "-archive/photo.jpg", want: folder + "-archive/photo.jpg"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var checkedPath string
			authorizer, err := fileaccess.NewAuthorizer(requestStub(func(_ context.Context, subject string, request any, response any) error {
				if subject != rbac.CanReadPathRPCSubject {
					t.Fatalf("unexpected RPC subject %q", subject)
				}
				checkedPath = request.(rbac.CheckPathRequest).Path
				response.(*rbac.DecisionResponse).Allowed = true
				return nil
			}))
			if err != nil {
				t.Fatal(err)
			}
			service := &fileService{
				auth: authorizer,
				cfg: serviceConfig{
					ACLPolicy: fileaccess.ReadACLPolicy{
						ApplicationFolders: map[string][]string{"moments": {folder}},
						EnforcePaths:       []string{folder + "/private"},
					},
				},
			}
			if err := service.authorizeRead(t.Context(), fileaccess.Subject{Username: "alice", UID: "1001", Application: test.app}, test.path); err != nil {
				t.Fatalf("authorizeRead() error = %v", err)
			}
			if checkedPath != test.want {
				t.Fatalf("RBAC checked %q, want %q", checkedPath, test.want)
			}
		})
	}
}

// TestListingChecksWhitelistedFolderAndEnforcedEntries covers file-service-svc/FR-008.
func TestListingChecksWhitelistedFolderAndEnforcedEntries(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"photo.jpg", "private.jpg"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("image"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	var checkedPaths []string
	service := mustService(t, root, func(_ context.Context, subject string, request any, response any) error {
		if subject == auth.ValidateAccessTokenRPCSubject {
			*response.(*auth.ValidateAccessTokenResponse) = auth.ValidateAccessTokenResponse{Valid: true, Status: auth.StatusAuthenticated, Username: "alice", Application: "moments"}
			return nil
		}
		if subject == rbac.GetSubjectRolesRPCSubject {
			roles := response.(*rbac.GetSubjectRolesResponse)
			roles.UID = "1001"
			roles.Groups = []string{"sudo"}
			return nil
		}
		if subject != rbac.CanReadPathRPCSubject {
			t.Fatalf("unexpected RPC subject %q", subject)
		}
		checkedPaths = append(checkedPaths, request.(rbac.CheckPathRequest).Path)
		response.(*rbac.DecisionResponse).Allowed = true
		return nil
	})
	service.cfg.ACLPolicy = fileaccess.ReadACLPolicy{
		ApplicationFolders: map[string][]string{"moments": {root}},
		EnforcePaths:       []string{filepath.Join(root, "private.jpg")},
	}
	request := httptest.NewRequest(http.MethodGet, "/ls"+root, nil)
	request.AddCookie(&http.Cookie{Name: "lite-nas-at", Value: "token"})
	response := httptest.NewRecorder()
	service.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("listing status = %d, want %d", response.Code, http.StatusOK)
	}
	want := []string{root, filepath.Join(root, "private.jpg")}
	if !slices.Equal(checkedPaths, want) {
		t.Fatalf("RBAC paths = %v, want %v", checkedPaths, want)
	}
}
