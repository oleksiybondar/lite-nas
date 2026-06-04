package services

import (
	"context"

	"lite-nas/shared/messaging"
)

func requestSnapshot[T any, R any](
	ctx context.Context,
	client messaging.Client,
	subject string,
	request any,
	selectSnapshot func(R) T,
) (T, error) {
	var response R
	if err := client.Request(ctx, subject, request, &response); err != nil {
		var zero T
		return zero, err
	}

	return selectSnapshot(response), nil
}

func requestHistory[T any, R any](
	ctx context.Context,
	client messaging.Client,
	subject string,
	request any,
	selectItems func(R) []T,
) ([]T, error) {
	var response R
	if err := client.Request(ctx, subject, request, &response); err != nil {
		return nil, err
	}

	return selectItems(response), nil
}
