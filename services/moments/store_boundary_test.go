package main

import "testing"

// TestGalleryStoreRejectsRelativePath protects the configured state location.
func TestGalleryStoreRejectsRelativePath(t *testing.T) {
	_, err := openGalleryStore("moments.db")
	if err == nil {
		t.Fatal("relative database path was accepted")
	}
}

// TestGalleryStoreRejectsDirectoryPath reports an unusable state target.
func TestGalleryStoreRejectsDirectoryPath(t *testing.T) {
	_, err := openGalleryStore(t.TempDir())
	if err == nil {
		t.Fatal("directory was accepted as a database")
	}
}
