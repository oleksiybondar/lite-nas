package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// mustCaptionWorker starts one controlled script and registers its cleanup.
func mustCaptionWorker(t *testing.T, program string) *captionWorker {
	t.Helper()
	script := filepath.Join(t.TempDir(), "caption.py")
	if err := os.WriteFile(script, []byte(program), 0o600); err != nil {
		t.Fatal(err)
	}
	places := &cityIndex{captionScript: script}
	worker := startCaptionWorker(context.Background(), places)
	if worker == nil {
		t.Fatal("caption worker did not start")
	}
	t.Cleanup(worker.close)
	return worker
}

// TestCaptionWorkerConsumesAuthorizedImage covers ordinary and large camera files.
func TestCaptionWorkerConsumesAuthorizedImage(t *testing.T) {
	for _, test := range []struct {
		name string
		size int
	}{
		{name: "ordinary image", size: 11},
		{name: "large camera image", size: 40 << 20},
	} {
		t.Run(test.name, func(t *testing.T) {
			program := "import sys\nprint('{\"ready\":true}', flush=True)\nfor line in sys.stdin: print('{\"caption\":\"A dog on a sofa\"}', flush=True)\n"
			worker := mustCaptionWorker(t, program)
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				writer.Header().Set("Content-Type", "image/jpeg")
				_, _ = writer.Write(bytes.Repeat([]byte("x"), test.size))
			}))
			defer server.Close()
			caption := worker.caption(context.Background(), newFileServiceClient(server.URL), "token", "/photo.jpg")
			if caption == nil || *caption != "A dog on a sofa" {
				t.Fatalf("caption = %v", caption)
			}
		})
	}
}

// TestMissingCaptionWorkerLeavesMetadataAvailable covers the optional model path.
func TestMissingCaptionWorkerLeavesMetadataAvailable(t *testing.T) {
	places := &cityIndex{captionScript: filepath.Join(t.TempDir(), "missing-caption.py")}
	if worker := startCaptionWorker(context.Background(), places); worker != nil {
		worker.close()
		t.Fatal("missing caption script started a worker")
	}
}

// TestCaptionRequestUsesFileServiceGrant keeps the browser token out of background reads.
func TestCaptionRequestUsesFileServiceGrant(t *testing.T) {
	var authorization string
	client := fileServiceClient{origin: "http://127.0.0.1:8082", client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		authorization = request.Header.Get("Authorization")
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("image"))}, nil
	})}}
	response, err := client.captionRequest(t.Context(), "grant", "/photos/lizard.jpg")
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if authorization != "CaptionGrant grant" {
		t.Fatalf("caption authorization = %q", authorization)
	}
}

// TestCaptionWorkerReadTimeout ensures a stalled model cannot hold a scan indefinitely.
func TestCaptionWorkerReadTimeout(t *testing.T) {
	program := "import time\nprint('{\"ready\":true}', flush=True)\ntime.sleep(60)\n"
	worker := mustCaptionWorker(t, program)
	if _, err := worker.readLine(context.Background(), 20*time.Millisecond); err == nil {
		t.Fatal("stalled caption worker did not time out")
	}
}
