package controllers

import (
	"context"
	"time"

	"github.com/danielgtaylor/huma/v2"
)

func fetchSnapshotOutput[T any, O any](
	ctx context.Context,
	fetch func(context.Context) (T, error),
	build func(time.Time, T) O,
	errorMessage string,
) (*O, error) {
	now := time.Now()
	snapshot, err := fetch(ctx)
	if err != nil {
		return nil, huma.Error502BadGateway(errorMessage)
	}

	output := build(now, snapshot)
	return &output, nil
}

func fetchHistoryOutput[T any, O any](
	ctx context.Context,
	fetch func(context.Context) ([]T, error),
	build func(time.Time, []T) O,
	errorMessage string,
) (*O, error) {
	now := time.Now()
	history, err := fetch(ctx)
	if err != nil {
		return nil, huma.Error502BadGateway(errorMessage)
	}

	output := build(now, history)
	return &output, nil
}
