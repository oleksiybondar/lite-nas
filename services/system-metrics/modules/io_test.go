package modules

import "testing"

func TestNewIOModuleCPUReaderReadsConfiguredFile(t *testing.T) {
	t.Parallel()

	module := loadIOModuleFixture(t)
	cpuData, err := module.CPUReader().Read()
	if err != nil {
		t.Fatalf("CPUReader().Read() error = %v", err)
	}

	if string(cpuData) != "cpu data" {
		t.Fatalf("CPUReader().Read() = %q, want %q", string(cpuData), "cpu data")
	}
}

func TestNewIOModuleMemReaderReadsConfiguredFile(t *testing.T) {
	t.Parallel()

	module := loadIOModuleFixture(t)
	memData, err := module.MemReader().Read()
	if err != nil {
		t.Fatalf("MemReader().Read() error = %v", err)
	}

	if string(memData) != "mem data" {
		t.Fatalf("MemReader().Read() = %q, want %q", string(memData), "mem data")
	}
}

func TestNewIOModuleReturnsReaderError(t *testing.T) {
	t.Parallel()

	_, err := NewIOModule("", "/missing-mem")
	if err == nil {
		t.Fatal("expected reader error")
	}
}
