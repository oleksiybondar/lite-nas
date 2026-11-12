package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

const maxEmbeddingWorkerLine = 48 << 20

type embeddingWork struct {
	id      uint64
	request []byte
}

type embeddingResult struct {
	id   uint64
	line []byte
	err  error
}

// embeddingWorker owns the resident Python model process. Its bounded input
// queue is fed by request callers; its bounded output queue is produced by the
// process bridge. Requests are serialized because the child has one stdin and
// stdout stream. The bridge owns neither channel's closure; shutdown cancels
// the process and leaves buffered replies accounted for by waiting callers.
type embeddingWorker struct {
	command  *exec.Cmd
	input    io.WriteCloser
	lines    *bufio.Reader
	in       chan embeddingWork
	out      chan embeddingResult
	done     chan struct{}
	mu       sync.Mutex
	nextID   uint64
	stop     chan struct{}
	stopOnce sync.Once
}

// startEmbeddingWorker starts and warms the selected model before serving APIs.
func startEmbeddingWorker(ctx context.Context, places *cityIndex) (*embeddingWorker, error) {
	if places == nil || places.embeddingScript == "" {
		return nil, errors.New("embedding model is disabled")
	}
	if _, err := os.Stat(places.embeddingScript); err != nil {
		return nil, err
	}
	if _, err := os.Stat(embeddingPython); err != nil {
		return nil, err
	}
	command := exec.CommandContext(ctx, embeddingPython, "-u", places.embeddingScript, "--serve", places.embeddingModel)
	input, err := command.StdinPipe()
	if err != nil {
		return nil, err
	}
	output, err := command.StdoutPipe()
	if err != nil {
		_ = input.Close()
		return nil, err
	}
	command.Stderr = os.Stderr
	if err := command.Start(); err != nil {
		_ = input.Close()
		return nil, fmt.Errorf("start embedding model worker: %w", err)
	}
	worker := &embeddingWorker{
		command: command,
		input:   input,
		lines:   bufio.NewReader(output),
		in:      make(chan embeddingWork, 1),
		out:     make(chan embeddingResult, 1),
		done:    make(chan struct{}),
		stop:    make(chan struct{}),
	}
	line, err := worker.readLine(ctx, 180*time.Second)
	var ready struct {
		Ready bool `json:"ready"`
	}
	if err != nil || json.Unmarshal(line, &ready) != nil || !ready.Ready {
		_ = worker.command.Process.Kill()
		_ = worker.input.Close()
		_ = worker.command.Wait()
		if err == nil {
			err = errors.New("embedding worker did not report ready")
		}
		return nil, err
	}
	go worker.run(ctx)
	return worker, nil
}

// run owns the subprocess streams and transfers bounded requests and replies.
func (worker *embeddingWorker) run(ctx context.Context) {
	defer close(worker.done)
	for {
		select {
		case <-ctx.Done():
			return
		case <-worker.stop:
			return
		case work := <-worker.in:
			if _, err := worker.input.Write(append(work.request, '\n')); err != nil {
				_ = worker.command.Process.Kill()
				worker.publish(ctx, embeddingResult{id: work.id, err: err})
				return
			}
			line, err := worker.readLine(ctx, 300*time.Second)
			worker.publish(ctx, embeddingResult{id: work.id, line: line, err: err})
			if err != nil {
				_ = worker.command.Process.Kill()
				return
			}
		}
	}
}

// publish sends one process reply through the bounded output channel.
func (worker *embeddingWorker) publish(ctx context.Context, result embeddingResult) {
	select {
	case worker.out <- result:
	case <-ctx.Done():
	}
}

// embed submits one request through the worker's bounded input/output queues.
func (worker *embeddingWorker) embed(ctx context.Context, model string, texts []string, size int) ([][]float32, error) {
	request, err := json.Marshal(struct {
		Texts []string `json:"texts"`
		Model string   `json:"model"`
	}{Texts: texts, Model: model})
	if err != nil {
		return nil, err
	}
	worker.mu.Lock()
	defer worker.mu.Unlock()
	worker.nextID++
	id := worker.nextID
	select {
	case worker.in <- embeddingWork{id: id, request: request}:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-worker.done:
		return nil, errors.New("embedding worker stopped")
	}
	for {
		select {
		case result := <-worker.out:
			if result.id != id {
				continue
			}
			if result.err != nil {
				return nil, fmt.Errorf("embedding worker failed: %w", result.err)
			}
			return decodeEmbeddingReply(result.line, len(texts), size)
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-worker.done:
			return nil, errors.New("embedding worker stopped")
		}
	}
}

// readLine bounds a single response and lets caller deadlines terminate a hung model.
func (worker *embeddingWorker) readLine(ctx context.Context, timeout time.Duration) ([]byte, error) {
	type readResult struct {
		line []byte
		err  error
	}
	result := make(chan readResult, 1)
	go func() {
		var line []byte
		var err error
		for {
			var part []byte
			part, err = worker.lines.ReadSlice('\n')
			line = append(line, part...)
			if len(line) > maxEmbeddingWorkerLine {
				err = errors.New("embedding worker output too large")
				break
			}
			if !errors.Is(err, bufio.ErrBufferFull) {
				break
			}
		}
		select {
		case result <- readResult{line: line, err: err}:
		case <-ctx.Done():
		}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case reply := <-result:
		return reply.line, reply.err
	case <-timer.C:
		_ = worker.command.Process.Kill()
		return nil, errors.New("embedding worker timed out")
	case <-ctx.Done():
		_ = worker.command.Process.Kill()
		return nil, ctx.Err()
	}
}

// close stops the resident worker during service shutdown.
func (worker *embeddingWorker) close() {
	worker.stopOnce.Do(func() { close(worker.stop) })
	_ = worker.input.Close()
	_ = worker.command.Wait()
	<-worker.done
}
