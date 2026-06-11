package controllers

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func assertWrappedData[T any](t *testing.T, got T, want T) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Data = %#v, want %#v", got, want)
	}
}

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

func assertBackendFailureMapped[T any](t *testing.T, got *T, err error, methodName string) {
	t.Helper()

	if err == nil {
		t.Fatalf("%s error = nil, want error", methodName)
	}

	if got != nil {
		t.Fatalf("%s result = %#v, want nil", methodName, got)
	}
}

func backendFailure() error {
	return errors.New("backend failed")
}

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
