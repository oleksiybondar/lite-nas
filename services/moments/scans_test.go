package main

import "testing"

// TestLivePhotoPairingUsesSameDirectoryAndBasename covers moments-backend/FR-004.
func TestLivePhotoPairingUsesSameDirectoryAndBasename(t *testing.T) {
	for _, test := range []struct {
		name, imagePath, moviePath, want string
	}{
		{"matched", "/home/alice/a.heic", "/home/alice/a.mov", "LIVE_PHOTO"},
		{"different directory", "/home/alice/a.heic", "/home/bob/a.mov", "IMAGE"},
		{"different case", "/home/alice/A.heic", "/home/alice/a.mov", "IMAGE"},
	} {
		t.Run(test.name, func(t *testing.T) {
			media := []scanMedia{{Path: test.imagePath, Type: "IMAGE"}, {Path: test.moviePath, Type: "VIDEO"}}
			pairLivePhotos(media)
			if media[0].Type != test.want {
				t.Fatalf("type = %q, want %q", media[0].Type, test.want)
			}
		})
	}
}
