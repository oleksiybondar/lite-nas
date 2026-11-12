package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	contract "lite-nas/shared/contracts/backuprestore"
	"lite-nas/shared/messaging"
)

const (
	mergeChunkSize = 64 << 10
	mergeFileLimit = 1 << 20
)

// runMerge handles start, resume, and discard without reading privileged files.
func runMerge(ctx context.Context, client messaging.Client, token string, args []string) error {
	if len(args) == 2 && validApplication(args[0]) && validVersion(args[1]) {
		var response contract.MergeStartResponse
		if err := client.Request(ctx, contract.MergeStartSubject, contract.MergeStartRequest{AccessToken: token, Application: args[0], Version: args[1]}, &response); err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, "merge session", response.SessionID)
		return editMerge(ctx, client, token, response)
	}
	if len(args) == 2 && args[0] == "resume" && validSessionID(args[1]) {
		var response contract.MergeStartResponse
		if err := client.Request(ctx, contract.MergeStatusSubject, contract.MergeActionRequest{AccessToken: token, SessionID: args[1]}, &response); err != nil {
			return err
		}
		return editMerge(ctx, client, token, response)
	}
	if len(args) == 2 && args[0] == "discard" && validSessionID(args[1]) {
		var response contract.MergeActionResponse
		if err := client.Request(ctx, contract.MergeDiscardSubject, contract.MergeActionRequest{AccessToken: token, SessionID: args[1]}, &response); err != nil {
			return err
		}
		if !response.Completed {
			return fmt.Errorf("discard was not confirmed")
		}
		return nil
	}
	return fmt.Errorf("usage: merge APP VERSION | merge resume SESSION_ID | merge discard SESSION_ID")
}

// validSessionID accepts the service-generated 128-bit hexadecimal identifier.
func validSessionID(id string) bool {
	if len(id) != 32 {
		return false
	}
	_, err := hex.DecodeString(id)
	return err == nil
}

// editMerge stages private local copies and submits every edited result in chunks.
func editMerge(ctx context.Context, client messaging.Client, token string, session contract.MergeStartResponse) error {
	if !validSessionID(session.SessionID) || len(session.Files) == 0 || len(session.Files) > 32 {
		return fmt.Errorf("invalid merge session response")
	}
	root, err := os.MkdirTemp("/tmp", "backup-restore-merge-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(root)
	if err := os.Chmod(root, 0o700); err != nil {
		return err
	}
	for index, file := range session.Files {
		if !filepath.IsAbs(file.Path) || filepath.Clean(file.Path) != file.Path {
			return fmt.Errorf("invalid merge path")
		}
		dir := filepath.Join(root, fmt.Sprintf("file-%d", index))
		if err := os.Mkdir(dir, 0o700); err != nil {
			return err
		}
		paths := map[string]string{"snapshot": filepath.Join(dir, "snapshot"), "current": filepath.Join(dir, "current"), "result": filepath.Join(dir, "result")}
		for _, side := range []string{"snapshot", "current", "result"} {
			content, err := downloadMerge(ctx, client, token, session.SessionID, file.Path, side)
			if err != nil {
				return err
			}
			mode := os.FileMode(0o400)
			if side == "result" {
				mode = 0o600
			}
			if err := os.WriteFile(paths[side], content, mode); err != nil {
				return err
			}
		}
		fmt.Fprintln(os.Stdout, "editing", file.Path, file.Status, "(snapshot, result, current)")
		if err := launchEditor(ctx, paths["snapshot"], paths["result"], paths["current"]); err != nil {
			return fmt.Errorf("merge session %s remains active: %w", session.SessionID, err)
		}
		result, err := readEditedResult(paths["result"])
		if err != nil {
			return err
		}
		if err := uploadMerge(ctx, client, token, session.SessionID, file.Path, result); err != nil {
			return err
		}
	}
	return finishMerge(ctx, client, token, session.SessionID, session.ApplyAvailable)
}

// launchEditor invokes one editor executable with three separate file operands.
func launchEditor(ctx context.Context, snapshot, result, current string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	if strings.HasPrefix(editor, "/usr/bin/") {
		editor = strings.TrimPrefix(editor, "/usr/bin/")
	}
	switch editor {
	case "vi", "vim", "nvim", "nano":
	default:
		return fmt.Errorf("EDITOR must be vi, vim, nvim, or nano")
	}
	command := exec.CommandContext(ctx, "/usr/bin/"+editor, snapshot, result, current)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	return command.Run()
}

// readEditedResult limits local result files before they cross NATS.
func readEditedResult(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > mergeFileLimit {
		return nil, fmt.Errorf("edited result is not a bounded regular file")
	}
	result, err := io.ReadAll(io.LimitReader(file, mergeFileLimit+1))
	if err != nil {
		return nil, err
	}
	if len(result) > mergeFileLimit {
		return nil, fmt.Errorf("edited result too large")
	}
	return result, nil
}

// downloadMerge checks each offset and the whole-content checksum.
func downloadMerge(ctx context.Context, client messaging.Client, token, id, path, side string) ([]byte, error) {
	var content []byte
	for {
		var response contract.MergeReadResponse
		request := contract.MergeReadRequest{AccessToken: token, SessionID: id, Path: path, Side: side, Offset: len(content)}
		if err := client.Request(ctx, contract.MergeReadSubject, request, &response); err != nil {
			return nil, err
		}
		chunk, err := base64.StdEncoding.DecodeString(response.Data)
		if err != nil || len(chunk) > mergeChunkSize || len(content)+len(chunk) > mergeFileLimit || response.NextOffset != len(content)+len(chunk) {
			return nil, fmt.Errorf("invalid merge download chunk")
		}
		content = append(content, chunk...)
		if response.Done {
			checksum := sha256.Sum256(content)
			if response.SHA256 != hex.EncodeToString(checksum[:]) {
				return nil, fmt.Errorf("merge download checksum mismatch")
			}
			return content, nil
		}
		if len(chunk) == 0 {
			return nil, fmt.Errorf("empty nonterminal merge chunk")
		}
	}
}

// uploadMerge sends the entire edited result sequentially with a final checksum.
func uploadMerge(ctx context.Context, client messaging.Client, token, id, path string, content []byte) error {
	checksum := sha256.Sum256(content)
	for offset := 0; ; {
		end := min(offset+mergeChunkSize, len(content))
		final := end == len(content)
		request := contract.MergeWriteRequest{AccessToken: token, SessionID: id, Path: path, Offset: offset, Data: base64.StdEncoding.EncodeToString(content[offset:end]), Final: final}
		if final {
			request.SHA256 = hex.EncodeToString(checksum[:])
		}
		var response contract.MergeWriteResponse
		if err := client.Request(ctx, contract.MergeWriteSubject, request, &response); err != nil {
			return err
		}
		if response.NextOffset != end {
			return fmt.Errorf("merge upload was not confirmed")
		}
		if final {
			return nil
		}
		offset = end
	}
}

// finishMerge asks for one explicit terminal action after all edits are staged.
func finishMerge(ctx context.Context, client messaging.Client, token, id string, applyAvailable bool) error {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer tty.Close()
	prompt := "Apply, save, or discard [apply/save/discard]: "
	if !applyAvailable {
		prompt = "Apply is unavailable; save or discard [save/discard]: "
	}
	if _, err := fmt.Fprint(tty, prompt); err != nil {
		return err
	}
	choice, err := bufio.NewReader(tty).ReadString('\n')
	if err != nil {
		return err
	}
	return submitMergeChoice(ctx, client, token, id, strings.TrimSpace(choice), applyAvailable)
}

// submitMergeChoice sends one validated terminal action to the service.
func submitMergeChoice(ctx context.Context, client messaging.Client, token, id, choice string, applyAvailable bool) error {
	switch choice {
	case "apply", "discard":
		if choice == "apply" && !applyAvailable {
			return fmt.Errorf("apply is unavailable; session remains active")
		}
		subject := contract.MergeApplySubject
		if choice == "discard" {
			subject = contract.MergeDiscardSubject
		}
		var response contract.MergeActionResponse
		if err := client.Request(ctx, subject, contract.MergeActionRequest{AccessToken: token, SessionID: id}, &response); err != nil {
			return err
		}
		if !response.Completed {
			return fmt.Errorf("merge action not confirmed")
		}
		if len(response.Warnings) > 0 {
			return fmt.Errorf("merge applied, but restart failed: %v", response.Warnings)
		}
		return nil
	case "save":
		var response contract.BackupResponse
		if err := client.Request(ctx, contract.MergeSaveSubject, contract.MergeSaveRequest{AccessToken: token, SessionID: id}, &response); err != nil {
			return err
		}
		fmt.Fprintln(os.Stdout, response.Version)
		return nil
	default:
		return fmt.Errorf("merge session %s remains active; use merge resume or discard", id)
	}
}
