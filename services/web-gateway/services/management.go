package services

import (
	"context"
	"errors"

	servicecontract "lite-nas/shared/contracts/servicesmanager"
	sessioncontract "lite-nas/shared/contracts/sessionmanager"
	usbcontract "lite-nas/shared/contracts/usbmanager"
	"lite-nas/shared/messaging"
)

// ErrManagementActionFailed means a backend did not confirm a management command.
var ErrManagementActionFailed = errors.New("management action not completed")

// ServicesManagementService provides authenticated system service inventory and commands.
type ServicesManagementService interface {
	List(context.Context, string) ([]servicecontract.Service, error)
	Action(context.Context, string, string, string) error
}

// SessionsManagementService provides authenticated host session inventory and commands.
type SessionsManagementService interface {
	List(context.Context, string) ([]sessioncontract.Session, error)
	Action(context.Context, string, string, string) error
}

// USBManagementService provides authenticated USBGuard inventory and authorization commands.
type USBManagementService interface {
	List(context.Context, string) (usbcontract.ListResponse, error)
	Action(context.Context, string, string, string, bool) error
}

type (
	servicesManagementService struct{ client messaging.Client }
	sessionsManagementService struct{ client messaging.Client }
	usbManagementService      struct{ client messaging.Client }
)

// NewServicesManagementService binds service management to the shared messaging client.
func NewServicesManagementService(client messaging.Client) ServicesManagementService {
	return servicesManagementService{client}
}

// NewSessionsManagementService binds session management to the shared messaging client.
func NewSessionsManagementService(client messaging.Client) SessionsManagementService {
	return sessionsManagementService{client}
}

// NewUSBManagementService binds USBGuard management to the shared messaging client.
func NewUSBManagementService(client messaging.Client) USBManagementService {
	return usbManagementService{client}
}

// List requests the current systemd service inventory using the caller's token.
func (s servicesManagementService) List(ctx context.Context, token string) ([]servicecontract.Service, error) {
	var response servicecontract.ListResponse
	err := s.client.Request(ctx, servicecontract.ListSubject, servicecontract.ListRequest{AccessToken: token}, &response)
	return response.Services, err
}

// Action sends one supported systemd command and requires backend confirmation.
func (s servicesManagementService) Action(ctx context.Context, token, unit, action string) error {
	subjects := map[string]string{"start": servicecontract.StartSubject, "stop": servicecontract.StopSubject, "restart": servicecontract.RestartSubject, "enable": servicecontract.EnableSubject, "disable": servicecontract.DisableSubject}
	subject, ok := subjects[action]
	if !ok {
		return ErrManagementActionFailed
	}
	var response servicecontract.ActionResponse
	if err := s.client.Request(ctx, subject, servicecontract.ActionRequest{AccessToken: token, Unit: unit}, &response); err != nil {
		return err
	}
	if !response.Completed {
		return ErrManagementActionFailed
	}
	return nil
}

// List requests the current loginctl session inventory using the caller's token.
func (s sessionsManagementService) List(ctx context.Context, token string) ([]sessioncontract.Session, error) {
	var response sessioncontract.ListResponse
	err := s.client.Request(ctx, sessioncontract.ListSubject, sessioncontract.ListRequest{AccessToken: token}, &response)
	return response.Sessions, err
}

// Action sends one supported loginctl command and requires backend confirmation.
func (s sessionsManagementService) Action(ctx context.Context, token, session, action string) error {
	subjects := map[string]string{"terminate": sessioncontract.TerminateSubject, "lock": sessioncontract.LockSubject, "unlock": sessioncontract.UnlockSubject}
	subject, ok := subjects[action]
	if !ok {
		return ErrManagementActionFailed
	}
	var response sessioncontract.ActionResponse
	if err := s.client.Request(ctx, subject, sessioncontract.ActionRequest{AccessToken: token, Session: session}, &response); err != nil {
		return err
	}
	if !response.Completed {
		return ErrManagementActionFailed
	}
	return nil
}

// List requests USBGuard's current device inventory using the caller's token.
func (s usbManagementService) List(ctx context.Context, token string) (usbcontract.ListResponse, error) {
	var response usbcontract.ListResponse
	err := s.client.Request(ctx, usbcontract.ListSubject, usbcontract.ListRequest{AccessToken: token}, &response)
	return response, err
}

// Action sends one supported USBGuard device decision and requires backend confirmation.
func (s usbManagementService) Action(ctx context.Context, token, deviceID, action string, permanent bool) error {
	subjects := map[string]string{"allow": usbcontract.AllowSubject, "block": usbcontract.BlockSubject}
	subject, ok := subjects[action]
	if !ok {
		return ErrManagementActionFailed
	}
	var response usbcontract.ActionResponse
	if err := s.client.Request(ctx, subject, usbcontract.DeviceRequest{AccessToken: token, DeviceID: deviceID, Permanent: permanent}, &response); err != nil {
		return err
	}
	if !response.Completed {
		return ErrManagementActionFailed
	}
	return nil
}
