package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	captionModel = "Salesforce/blip-image-captioning-base"
	// Match the Python worker's compressed-image limit before base64 encoding.
	maxCaptionImageBytes = 64 << 20
)

// captionWorker keeps one optional local model loaded for a scan or upload.
type captionWorker struct {
	command       *exec.Cmd
	input         io.WriteCloser
	output        *bufio.Reader
	unlock        func()
	failed        bool
	grantRejected bool
}

// startCaptionWorker uses only a provisioned local script and cached model weights.
func startCaptionWorker(ctx context.Context, places *cityIndex) *captionWorker {
	if places == nil || places.captionScript == "" {
		return nil
	}
	if _, err := os.Stat(places.captionScript); err != nil {
		return nil
	}
	places.captionMu.Lock()
	command := exec.CommandContext(ctx, "python3", "-u", places.captionScript)
	input, err := command.StdinPipe()
	if err != nil {
		places.captionMu.Unlock()
		return nil
	}
	output, err := command.StdoutPipe()
	if err != nil {
		_ = input.Close()
		places.captionMu.Unlock()
		return nil
	}
	// Model import and AppArmor failures belong in the service journal.
	command.Stderr = os.Stderr
	if err := command.Start(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "moments: BLIP worker could not start: %v\n", err)
		_ = input.Close()
		places.captionMu.Unlock()
		return nil
	}
	worker := &captionWorker{command: command, input: input, output: bufio.NewReader(output), unlock: places.captionMu.Unlock}
	var ready struct {
		Ready bool `json:"ready"`
	}
	line, err := worker.readLine(ctx, 180*time.Second)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "moments: BLIP worker startup failed: %v\n", err)
		worker.close()
		return nil
	}
	if json.Unmarshal(line, &ready) != nil || !ready.Ready {
		_, _ = fmt.Fprintln(os.Stderr, "moments: BLIP worker did not report ready")
		worker.close()
		return nil
	}
	return worker
}

// caption streams one ACL-authorized image to the local model without storing its bytes.
func (worker *captionWorker) caption(ctx context.Context, files fileServiceClient, token, path string) *string {
	return worker.captionWithRequest(ctx, files, token, path, false)
}

// captionWithGrant reads one job-scoped source without a browser credential.
func (worker *captionWorker) captionWithGrant(ctx context.Context, files fileServiceClient, grant, path string) *string {
	return worker.captionWithRequest(ctx, files, grant, path, true)
}

// captionWithRequest streams one authorized image into the local model.
func (worker *captionWorker) captionWithRequest(ctx context.Context, files fileServiceClient, token, path string, grant bool) *string {
	if worker.failed {
		return nil
	}
	var response *http.Response
	var err error
	for {
		if grant {
			response, err = files.captionRequest(ctx, token, path)
		} else {
			response, err = files.request(ctx, token, "", path, "")
		}
		if !grant || (err == nil && response.StatusCode < http.StatusInternalServerError) {
			break
		}
		if response != nil {
			_ = response.Body.Close()
		}
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(10 * time.Second):
		}
	}
	if err != nil {
		return nil
	}
	defer response.Body.Close()
	if grant && response.StatusCode == http.StatusUnauthorized {
		worker.grantRejected = true
		return nil
	}
	if response.StatusCode != http.StatusOK || response.ContentLength > maxCaptionImageBytes {
		return nil
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxCaptionImageBytes+1))
	if err != nil || len(data) == 0 || len(data) > maxCaptionImageBytes {
		return nil
	}
	request, err := json.Marshal(struct {
		Image string `json:"image"`
	}{Image: base64.StdEncoding.EncodeToString(data)})
	if err != nil {
		return nil
	}
	if _, err := io.Copy(worker.input, bytes.NewReader(append(request, '\n'))); err != nil {
		worker.failed = true
		return nil
	}
	line, err := worker.readLine(ctx, 180*time.Second)
	if err != nil || len(line) > 4096 {
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "moments: BLIP image caption failed: %v\n", err)
		} else {
			_, _ = fmt.Fprintln(os.Stderr, "moments: BLIP image caption response exceeded limit")
		}
		worker.failed = true
		return nil
	}
	var result struct {
		Caption string `json:"caption"`
	}
	if json.Unmarshal(line, &result) != nil {
		worker.failed = true
		return nil
	}
	caption := strings.TrimSpace(result.Caption)
	if caption == "" || len(caption) > 512 {
		return nil
	}
	return &caption
}

// readLine bounds one local worker reply; its single producer sends to a one-slot
// result channel, which needs no close because the read ends after one reply.
func (worker *captionWorker) readLine(ctx context.Context, timeout time.Duration) ([]byte, error) {
	type result struct {
		line []byte
		err  error
	}
	replies := make(chan result, 1)
	go func() {
		line, err := worker.output.ReadSlice('\n')
		select {
		case replies <- result{line: line, err: err}:
		case <-ctx.Done():
		}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case reply := <-replies:
		if reply.err != nil {
			worker.failed = true
			_ = worker.command.Process.Kill()
		}
		return reply.line, reply.err
	case <-timer.C:
		worker.failed = true
		_ = worker.command.Process.Kill()
		return nil, errors.New("caption worker timed out")
	case <-ctx.Done():
		worker.failed = true
		_ = worker.command.Process.Kill()
		return nil, ctx.Err()
	}
}

// close stops the local process and releases the single-model capacity.
func (worker *captionWorker) close() {
	_ = worker.input.Close()
	_ = worker.command.Wait()
	worker.unlock()
}
