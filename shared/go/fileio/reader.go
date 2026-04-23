package fileio

import (
	"errors"
	"os"
)

// Reader defines a minimal interface for reading file contents.
//
// It abstracts file access to improve testability and decouple higher-level
// components from the underlying filesystem. Implementations may read from
// the OS, in-memory structures, or other sources.
type Reader interface {
	// Read returns the full contents of the configured file.
	//
	// If the file cannot be read, an error is returned.
	Read() ([]byte, error)
}

// FileReader is a Reader implementation bound to a single file path.
//
// The file path is provided at construction time so the dependency is explicit
// and file access remains controlled by the application.
type FileReader struct {
	FilePath string
}

// NewFileReader creates a FileReader bound to a specific file path.
//
// An error is returned if the provided path is empty.
func NewFileReader(path string) (FileReader, error) {
	if path == "" {
		return FileReader{}, errors.New("empty path")
	}

	return FileReader{FilePath: path}, nil
}

// Read reads the full contents of the configured file.
//
// The file path is fixed when the reader is created and is not supplied at
// call time.
func (r FileReader) Read() ([]byte, error) {
	return os.ReadFile(r.FilePath) // #nosec G304 -- path is configured by the application
}
