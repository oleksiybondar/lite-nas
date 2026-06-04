package services

import (
	"context"
	"errors"

	loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"
	securityloggingmanagercontract "lite-nas/shared/contracts/securityloggingmanager"
	systemloggingmanagercontract "lite-nas/shared/contracts/systemloggingmanager"
	"lite-nas/shared/messaging"
)

// ErrAlertActionFailed reports a backend alert mutation failure acknowledged by the RPC layer.
var ErrAlertActionFailed = errors.New("alert action failed")

// AlertsService defines the backend-facing alert flows used by one gateway alert domain.
type AlertsService interface {
	List(context.Context, AlertListInput) (AlertListPage, error)
	ListUnacknowledged(context.Context, AlertListInput) (AlertListPage, error)
	Get(context.Context, AlertGetInput) (loggingmanagercontract.ListAlertItem, bool, error)
	Acknowledge(context.Context, AlertActionInput) error
	Mute(context.Context, AlertActionInput) error
}

// AlertListInput defines one browser-facing alert-list request after controller normalization.
type AlertListInput struct {
	AccessToken string
	Page        int
	Size        int
}

// AlertGetInput defines one browser-facing alert-detail request after controller extraction.
type AlertGetInput struct {
	AccessToken string
	ID          string
}

// AlertActionInput defines one browser-facing alert command after controller extraction.
type AlertActionInput struct {
	AccessToken string
	ID          string
	ActorLogin  string
}

// AlertListPage contains one browser-facing alert page and its total count.
type AlertListPage struct {
	Items      []loggingmanagercontract.ListAlertItem
	TotalCount int
}

type alertSubjects struct {
	getAll            string
	getUnacknowledged string
	getOne            string
	acknowledge       string
	mute              string
}

type alertsService struct {
	client   messaging.Client
	subjects alertSubjects
}

// NewSystemAlertsService creates a gateway alert service bound to the system logging-manager RPC subjects.
func NewSystemAlertsService(client messaging.Client) AlertsService {
	return alertsService{
		client: client,
		subjects: alertSubjects{
			getAll:            systemloggingmanagercontract.GetAlertsRPCSubject,
			getUnacknowledged: systemloggingmanagercontract.GetUnacknowledgedActiveAlertsRPCSubject,
			getOne:            systemloggingmanagercontract.GetAlertRPCSubject,
			acknowledge:       systemloggingmanagercontract.AcknowledgeAlertRPCSubject,
			mute:              systemloggingmanagercontract.MuteAlertRPCSubject,
		},
	}
}

// NewSecurityAlertsService creates a gateway alert service bound to the security logging-manager RPC subjects.
func NewSecurityAlertsService(client messaging.Client) AlertsService {
	return alertsService{
		client: client,
		subjects: alertSubjects{
			getAll:            securityloggingmanagercontract.GetAlertsRPCSubject,
			getUnacknowledged: securityloggingmanagercontract.GetUnacknowledgedActiveAlertsRPCSubject,
			getOne:            securityloggingmanagercontract.GetAlertRPCSubject,
			acknowledge:       securityloggingmanagercontract.AcknowledgeAlertRPCSubject,
			mute:              securityloggingmanagercontract.MuteAlertRPCSubject,
		},
	}
}

// List requests one page of alerts from the configured logging-manager domain.
func (s alertsService) List(ctx context.Context, input AlertListInput) (AlertListPage, error) {
	return s.requestList(ctx, s.subjects.getAll, input)
}

// ListUnacknowledged requests one page of unacknowledged alerts from the configured logging-manager domain.
func (s alertsService) ListUnacknowledged(ctx context.Context, input AlertListInput) (AlertListPage, error) {
	return s.requestList(ctx, s.subjects.getUnacknowledged, input)
}

// Get requests one alert by business record ID from the configured logging-manager domain.
func (s alertsService) Get(ctx context.Context, input AlertGetInput) (loggingmanagercontract.ListAlertItem, bool, error) {
	var response loggingmanagercontract.GetAlertResponse
	request := loggingmanagercontract.GetAlertInput{
		AccessToken: input.AccessToken,
		EventID:     input.ID,
	}
	if err := s.client.Request(ctx, s.subjects.getOne, request, &response); err != nil {
		return loggingmanagercontract.ListAlertItem{}, false, err
	}
	if response.Item == nil {
		return loggingmanagercontract.ListAlertItem{}, false, nil
	}
	return *response.Item, true, nil
}

// Acknowledge requests alert acknowledgement in the configured logging-manager domain.
func (s alertsService) Acknowledge(ctx context.Context, input AlertActionInput) error {
	request := loggingmanagercontract.AcknowledgeAlertInput{
		AccessToken:    input.AccessToken,
		EventID:        input.ID,
		AcknowledgedBy: input.ActorLogin,
	}
	return s.requestAction(ctx, s.subjects.acknowledge, request)
}

// Mute requests alert muting in the configured logging-manager domain.
func (s alertsService) Mute(ctx context.Context, input AlertActionInput) error {
	request := loggingmanagercontract.MuteAlertInput{
		AccessToken: input.AccessToken,
		EventID:     input.ID,
		MutedBy:     input.ActorLogin,
	}
	return s.requestAction(ctx, s.subjects.mute, request)
}

func (s alertsService) requestList(ctx context.Context, subject string, input AlertListInput) (AlertListPage, error) {
	var response loggingmanagercontract.ListAlertsResponse
	request := loggingmanagercontract.ListAlertsInput{
		AccessToken: input.AccessToken,
		Page:        input.Page,
		PageSize:    input.Size,
	}
	if err := s.client.Request(ctx, subject, request, &response); err != nil {
		return AlertListPage{}, err
	}
	return AlertListPage{Items: response.Items, TotalCount: response.TotalCount}, nil
}

func (s alertsService) requestAction(ctx context.Context, subject string, request any) error {
	var response loggingmanagercontract.OKResponse
	if err := s.client.Request(ctx, subject, request, &response); err != nil {
		return err
	}
	if !response.OK {
		return ErrAlertActionFailed
	}
	return nil
}
