package modules

import (
	sharedfileio "lite-nas/shared/fileio"
)

// IO groups low-level readers used by the service runtime.
//
// The fields are populated once during startup and are expected to be treated
// as logically read-only after construction.
type IO struct {
	CPUReader sharedfileio.Reader
	MemReader sharedfileio.Reader
}

// NewIOModule opens the procfs readers required by the metrics workers.
//
// Parameters:
//   - cpuPath: path to the procfs CPU statistics source
//   - memPath: path to the procfs memory statistics source
func NewIOModule(cpuPath string, memPath string) (IO, error) {
	cpuReader, err := sharedfileio.NewFileReader(cpuPath)
	if err != nil {
		return IO{}, err
	}

	memReader, err := sharedfileio.NewFileReader(memPath)
	if err != nil {
		return IO{}, err
	}

	return IO{
		CPUReader: cpuReader,
		MemReader: memReader,
	}, nil
}
