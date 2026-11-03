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

// TestFriendListSearchPaginatesSavedContacts checks server-side filtering and page metadata.
func TestFriendListSearchPaginatesSavedContacts(t *testing.T) {
	router := mustGalleryRouter(t, "1001")
	mustGalleryRequest(t, router, http.MethodPut, "/api/friends/aria", `{"firstName":"Ari"}`, http.StatusNoContent)
	mustGalleryRequest(t, router, http.MethodPut, "/api/friends/bob", `{"firstName":"Ari"}`, http.StatusNoContent)
	mustGalleryRequest(t, router, http.MethodPut, "/api/friends/cora", `{"firstName":"Cora"}`, http.StatusNoContent)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/friends?q=ari&page=2&size=1", "", http.StatusOK)
	page := mustGalleryJSON[struct {
		Items    []friendRecord `json:"items"`
		Metadata friendPage     `json:"metadata"`
	}](t, body)
	if len(page.Items) != 1 || page.Items[0].Username != "bob" || page.Metadata != (friendPage{Page: 2, Size: 1, TotalCount: 2, TotalPages: 2}) {
		t.Fatalf("friend search page = %+v", page)
	}
}

// TestFriendListSearchTreatsWildcardsLiterally checks that search text cannot broaden a LIKE query.
func TestFriendListSearchTreatsWildcardsLiterally(t *testing.T) {
	router := mustGalleryRouter(t, "1001")
	mustGalleryRequest(t, router, http.MethodPut, "/api/friends/al_ice", `{}`, http.StatusNoContent)
	mustGalleryRequest(t, router, http.MethodPut, "/api/friends/alice", `{}`, http.StatusNoContent)
	body := mustGalleryRequest(t, router, http.MethodGet, "/api/friends?q=al_", "", http.StatusOK)
	page := mustGalleryJSON[struct {
		Items []friendRecord `json:"items"`
	}](t, body)
	if len(page.Items) != 1 || page.Items[0].Username != "al_ice" {
		t.Fatalf("literal wildcard search = %+v", page.Items)
	}
}
