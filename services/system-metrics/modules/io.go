package modules

import (
	sharedfileio "lite-nas/shared/fileio"
)

// IO groups low-level reader dependencies used by the service.
type IO struct {
	cpuReader sharedfileio.Reader
	memReader sharedfileio.Reader
}

// NewIOModule creates the reader dependencies required by the metrics workers.
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
		cpuReader: cpuReader,
		memReader: memReader,
	}, nil
}

// CPUReader returns the CPU metrics reader.
func (m IO) CPUReader() sharedfileio.Reader {
	return m.cpuReader
}

// MemReader returns the memory metrics reader.
func (m IO) MemReader() sharedfileio.Reader {
	return m.memReader
}
