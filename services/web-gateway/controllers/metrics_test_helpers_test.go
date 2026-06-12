package controllers

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

// assertWrappedData verifies one DTO envelope data payload matches the expected snapshot value.
func assertWrappedData[T any](t *testing.T, got T, want T) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Data = %#v, want %#v", got, want)
	}
}

// assertWrappedSlice verifies one DTO envelope data payload matches the expected history values.
func assertWrappedSlice[T any](t *testing.T, got []T, want []T) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("len(Data) = %d, want %d", len(got), len(want))
	}

	for i := range want {
		if !reflect.DeepEqual(got[i], want[i]) {
			t.Fatalf("Data[%d] = %#v, want %#v", i, got[i], want[i])
		}
	}
}

// assertBackendFailureMapped verifies one controller method maps a backend failure into an error result.
func assertBackendFailureMapped[T any](t *testing.T, got *T, err error, methodName string) {
	t.Helper()

	if err == nil {
		t.Fatalf("%s error = nil, want error", methodName)
	}

	if got != nil {
		t.Fatalf("%s result = %#v, want nil", methodName, got)
	}
}

// backendFailure returns one stable backend error for controller failure mapping tests.
func backendFailure() error {
	return errors.New("backend failed")
}

// assertSnapshotWrapped verifies one snapshot endpoint returns a successful envelope and expected data.
func assertSnapshotWrapped[T any, O any](
	t *testing.T,
	snapshot T,
	getSnapshot func(context.Context, *struct{}) (*O, error),
	success func(*O) bool,
	timestampIsZero func(*O) bool,
	data func(*O) T,
) {
	t.Helper()

	got, err := getSnapshot(context.Background(), &struct{}{})
	if err != nil {
		t.Fatalf("GetSnapshot() error = %v", err)
	}

	assertSuccessfulSystemMetricsEnvelope(t, success(got), timestampIsZero(got))
	assertWrappedData(t, data(got), snapshot)
}

// assertHistoryWrapped verifies one history endpoint returns a successful envelope and expected data.
func assertHistoryWrapped[T any, O any](
	t *testing.T,
	history []T,
	getHistory func(context.Context, *struct{}) (*O, error),
	success func(*O) bool,
	timestampIsZero func(*O) bool,
	data func(*O) []T,
) {
	t.Helper()

	got, err := getHistory(context.Background(), &struct{}{})
	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}

	assertSuccessfulSystemMetricsEnvelope(t, success(got), timestampIsZero(got))
	assertWrappedSlice(t, data(got), history)
}

// runSnapshotEnvelopeTest constructs one controller and verifies its snapshot envelope response.
func runSnapshotEnvelopeTest[T any, O any](
	t *testing.T,
	snapshot T,
	newGetSnapshot func(T) func(context.Context, *struct{}) (*O, error),
	success func(*O) bool,
	timestampIsZero func(*O) bool,
	data func(*O) T,
) {
	t.Helper()

	assertSnapshotWrapped(t, snapshot, newGetSnapshot(snapshot), success, timestampIsZero, data)
}

// runHistoryEnvelopeTest constructs one controller and verifies its history envelope response.
func runHistoryEnvelopeTest[T any, O any](
	t *testing.T,
	history []T,
	newGetHistory func([]T) func(context.Context, *struct{}) (*O, error),
	success func(*O) bool,
	timestampIsZero func(*O) bool,
	data func(*O) []T,
) {
	t.Helper()

	assertHistoryWrapped(t, history, newGetHistory(history), success, timestampIsZero, data)
}
