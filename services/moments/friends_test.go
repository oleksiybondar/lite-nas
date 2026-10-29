package main

import (
	"net/http"
	"testing"
)

// TestFriendListAcceptsUnknownUnixLogins verifies moments-backend/FR-010 without account lookup.
func TestFriendListAcceptsUnknownUnixLogins(t *testing.T) {
	router := mustGalleryRouter(t, "1001")
	mustGalleryRequest(t, router, http.MethodPut, "/api/friends/not-a-real-account", `{"firstName":"Not","lastName":"Found","nickname":"imaginary"}`, http.StatusNoContent)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/friends", "", http.StatusOK)
	friends := mustGalleryJSON[struct {
		Items []friendRecord `json:"items"`
	}](t, body)
	if len(friends.Items) != 1 || friends.Items[0].Username != "not-a-real-account" || friends.Items[0].Nickname != "imaginary" {
		t.Fatalf("saved friend = %+v", friends.Items)
	}
}

// TestFriendListRejectsMalformedUnixLogin validates the moments-backend/FR-010 login boundary.
func TestFriendListRejectsMalformedUnixLogin(t *testing.T) {
	router := mustGalleryRouter(t, "1001")
	mustGalleryRequest(t, router, http.MethodPut, "/api/friends/--help", `{}`, http.StatusUnprocessableEntity)
}

// TestFriendListRemovalDeletesSavedLogin covers moments-backend/FR-010 contact removal.
func TestFriendListRemovalDeletesSavedLogin(t *testing.T) {
	router := mustGalleryRouter(t, "1001")
	mustGalleryRequest(t, router, http.MethodPut, "/api/friends/bob", `{}`, http.StatusNoContent)
	mustGalleryRequest(t, router, http.MethodDelete, "/api/friends/bob", "", http.StatusNoContent)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/friends", "", http.StatusOK)
	friends := mustGalleryJSON[struct {
		Items []friendRecord `json:"items"`
	}](t, body)
	if len(friends.Items) != 0 {
		t.Fatalf("friends after removal = %+v", friends.Items)
	}
}
