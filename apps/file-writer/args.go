package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/go-playground/validator/v10"
)

// helpText describes the writer's filesystem operations and identity parameters.
const helpText = `lite-nas-file-writer performs one upload-stage operation as the current Unix user.

Usage:
  lite-nas-file-writer create   --target PATH
  lite-nas-file-writer write    --target PATH --expected-device N --expected-inode N --expected-uid N --offset N --length N
  lite-nas-file-writer stat     --target PATH --expected-device N --expected-inode N --expected-uid N
  lite-nas-file-writer checksum --target PATH --expected-device N --expected-inode N --expected-uid N
  lite-nas-file-writer truncate --target PATH --expected-device N --expected-inode N --expected-uid N --offset N
  lite-nas-file-writer publish  --target PATH --expected-device N --expected-inode N --expected-uid N --destination-absent [--mode 0644|0660|0664]
  lite-nas-file-writer publish  --target PATH --expected-device N --expected-inode N --expected-uid N --destination-device N --destination-inode N [--mode 0644|0660|0664]
  lite-nas-file-writer cleanup  --target PATH --expected-device N --expected-inode N --expected-uid N

PATH is an absolute, clean path ending in .upload. The destination is PATH
without that suffix. A create fails if the stage already exists. File Service
owns authorization, upload metadata, and the expected identity passed to later
operations. The writer rejects a stage whose device, inode, or owner differs.

Write reads exactly --length bytes from standard input (at most 16 MiB),
replaces any uncommitted tail from --offset, and syncs before reporting success.
Truncate discards bytes after --offset. Checksum reports SHA-256 of the stage.
Publish atomically renames the stage into its destination after checking the
destination expectation; a failed publish attempts to remove the stage.
Cleanup only removes the identified .upload stage.

Success is one JSON object on standard output. Errors are JSON on standard
error with a nonzero exit status. Use --help or -h to print this message.
`

// invocation is the validated command input passed to one filesystem operation.
type invocation struct {
	Command           string `validate:"oneof=create write stat checksum truncate publish cleanup"`
	Target            string `validate:"required,max=4096"`
	Offset            int64  `validate:"min=-1"`
	Length            int64  `validate:"min=-1,max=16777216"`
	ExpectedDevice    uint64
	ExpectedInode     uint64
	ExpectedUID       int64 `validate:"min=-1"`
	DestinationAbsent bool
	DestinationDevice uint64
	DestinationInode  uint64
	Mode              string
}

// parseInvocation validates both the shared CLI shape and command-specific flags.
func parseInvocation(args []string) (invocation, error) {
	if len(args) == 0 {
		return invocation{}, errors.New("missing command")
	}
	inv := invocation{Command: args[0], Offset: -1, Length: -1, ExpectedUID: -1}
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	flags.StringVar(&inv.Target, "target", "", "absolute .upload stage path")
	flags.Int64Var(&inv.Offset, "offset", -1, "committed byte offset")
	flags.Int64Var(&inv.Length, "length", -1, "exact stdin byte count")
	flags.Uint64Var(&inv.ExpectedDevice, "expected-device", 0, "stage device number")
	flags.Uint64Var(&inv.ExpectedInode, "expected-inode", 0, "stage inode number")
	flags.Int64Var(&inv.ExpectedUID, "expected-uid", -1, "stage owner UID")
	flags.BoolVar(&inv.DestinationAbsent, "destination-absent", false, "require an absent destination")
	flags.Uint64Var(&inv.DestinationDevice, "destination-device", 0, "existing destination device")
	flags.Uint64Var(&inv.DestinationInode, "destination-inode", 0, "existing destination inode")
	flags.StringVar(&inv.Mode, "mode", "0664", "published file mode: 0644, 0660 or 0664")
	if err := flags.Parse(args[1:]); err != nil {
		return invocation{}, err
	}
	if err := validator.New().Struct(inv); err != nil {
		return invocation{}, fmt.Errorf("invalid arguments: %w", err)
	}
	if flags.NArg() != 0 || !filepath.IsAbs(inv.Target) || filepath.Clean(inv.Target) != inv.Target ||
		!strings.HasSuffix(inv.Target, ".upload") || filepath.Base(inv.Target) == ".upload" {
		return invocation{}, errors.New("target must be an absolute, clean .upload path")
	}
	allowed := map[string]bool{"target": true}
	if inv.Command != "create" {
		allowed["expected-device"] = true
		allowed["expected-inode"] = true
		allowed["expected-uid"] = true
		if inv.ExpectedDevice == 0 || inv.ExpectedInode == 0 || inv.ExpectedUID < 0 {
			return invocation{}, errors.New("expected-device, expected-inode, and expected-uid are required")
		}
	}
	switch inv.Command {
	case "write":
		allowed["offset"] = true
		allowed["length"] = true
		if inv.Offset < 0 || inv.Length < 0 {
			return invocation{}, errors.New("write requires --offset and --length")
		}
	case "truncate":
		allowed["offset"] = true
		if inv.Offset < 0 {
			return invocation{}, errors.New("truncate requires --offset")
		}
	case "publish":
		allowed["mode"] = true
		if inv.Mode != "0644" && inv.Mode != "0660" && inv.Mode != "0664" {
			return invocation{}, errors.New("publish mode must be 0644, 0660 or 0664")
		}
		allowed["destination-absent"] = true
		allowed["destination-device"] = true
		allowed["destination-inode"] = true
		if inv.DestinationAbsent == (inv.DestinationDevice != 0 && inv.DestinationInode != 0) ||
			(inv.DestinationDevice == 0) != (inv.DestinationInode == 0) {
			return invocation{}, errors.New("publish requires either --destination-absent or both destination identity flags")
		}
	}
	var disallowed string
	flags.Visit(func(value *flag.Flag) {
		if !allowed[value.Name] {
			disallowed = value.Name
		}
	})
	if disallowed != "" {
		return invocation{}, fmt.Errorf("--%s is not valid for %s", disallowed, inv.Command)
	}
	return inv, nil
}
