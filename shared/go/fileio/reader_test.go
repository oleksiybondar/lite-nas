package fileio_test

import (
	"os"
	"path/filepath"
	"testing"

	"lite-nas/shared/fileio"
)

func TestNewFileReaderRejectsEmptyPath(t *testing.T) {
	t.Parallel()

	if _, err := fileio.NewFileReader(""); err == nil {
		t.Fatal("expected empty path error")
	}
}

func TestFileReaderReadReturnsFileContents(t *testing.T) {
	t.Parallel()

	data := loadFileContentsFixture(t)
	if string(data) != "payload" {
		t.Fatalf("Read() = %q, want payload", string(data))
	}
}

func loadFileContentsFixture(t *testing.T) []byte {
	t.Helper()

	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, "metrics.txt")

	if err := os.WriteFile(filePath, []byte("payload"), 0o600); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	reader, err := fileio.NewFileReader(filePath)
	if err != nil {
		t.Fatalf("NewFileReader() error = %v", err)
	}

	data, err := reader.Read()
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	return data
}
