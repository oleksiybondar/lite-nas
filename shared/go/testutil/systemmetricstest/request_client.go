package systemmetricstest

import (
	"context"
	"errors"

	systemmetricscontract "lite-nas/shared/contracts/systemmetrics"
	"lite-nas/shared/metrics"
)

// RequestClient is a messaging.Client test double for system-metrics request/
// reply flows. It records the last subject and request payload while serving a
// configured contract response.
type RequestClient struct {
	Subject     string
	LastRequest any
	Response    any
	Err         error
}

func NewSnapshotClient(snapshot metrics.SystemSnapshot) *RequestClient {
	return &RequestClient{
		Response: systemmetricscontract.GetSnapshotResponse{
			Available: true,
			Snapshot:  snapshot,
		},
	}
}

func NewInactiveSnapshotClient() *RequestClient {
	return &RequestClient{
		Response: systemmetricscontract.GetSnapshotResponse{
			Available: false,
		},
	}
}

func NewHistoryClient(items []metrics.SystemSnapshot) *RequestClient {
	return &RequestClient{
		Response: systemmetricscontract.GetHistoryResponse{
			Items: items,
		},
	}
}

func NewRequestErrorClient(err error) *RequestClient {
	return &RequestClient{Err: err}
}

func (c *RequestClient) Publish(context.Context, string, any) error {
	return nil
}

func (c *RequestClient) Request(
	_ context.Context,
	subject string,
	request any,
	response any,
) error {
	c.Subject = subject
	c.LastRequest = request

	if c.Err != nil {
		return c.Err
	}

	switch payload := c.Response.(type) {
	case systemmetricscontract.GetSnapshotResponse:
		target, ok := response.(*systemmetricscontract.GetSnapshotResponse)
		if !ok {
			return errors.New("unexpected snapshot response target")
		}

		*target = payload
	case systemmetricscontract.GetHistoryResponse:
		target, ok := response.(*systemmetricscontract.GetHistoryResponse)
		if !ok {
			return errors.New("unexpected history response target")
		}

		*target = payload
	default:
		return errors.New("unexpected stub response type")
	}

	return nil
}

func (c *RequestClient) Drain() error {
	return nil
}

func (c *RequestClient) Close() {}
