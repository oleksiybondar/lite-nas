package controllers

import (
	"context"
	"testing"

	"lite-nas/services/web-gateway/dto/management"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/shared/authtoken"
	servicecontract "lite-nas/shared/contracts/servicesmanager"
	sessioncontract "lite-nas/shared/contracts/sessionmanager"
	usbcontract "lite-nas/shared/contracts/usbmanager"
)

type managementServiceContractStub struct{ token string }

func (s *managementServiceContractStub) List(_ context.Context, token string) ([]servicecontract.Service, error) {
	s.token = token
	return []servicecontract.Service{{Name: "ssh.service"}}, nil
}

func (s *managementServiceContractStub) Action(_ context.Context, token, _, _ string) error {
	s.token = token
	return nil
}

type managementSessionContractStub struct{ token string }

func (s *managementSessionContractStub) List(_ context.Context, token string) ([]sessioncontract.Session, error) {
	s.token = token
	return []sessioncontract.Session{{ID: "3"}}, nil
}

func (s *managementSessionContractStub) Action(_ context.Context, token, _, _ string) error {
	s.token = token
	return nil
}

type managementUSBContractStub struct{ token string }

func (s *managementUSBContractStub) List(_ context.Context, token string) (usbcontract.ListResponse, error) {
	s.token = token
	return usbcontract.ListResponse{}, nil
}

func (s *managementUSBContractStub) Action(_ context.Context, token, _, _ string, _ bool) error {
	s.token = token
	return nil
}

// TestManagementControllersForwardAccessToken verifies privileged host operations retain the caller's credential.
func TestManagementControllersForwardAccessToken(t *testing.T) {
	ctx := middlewares.NewAuthenticatedContext(context.Background(), "caller-token", authtoken.AccessClaims{})
	cases := []struct {
		name string
		call func() (string, error)
	}{
		{"services list", func() (string, error) {
			s := &managementServiceContractStub{}
			_, err := NewServicesManagementController(s).List(ctx, &management.ListInput{})
			return s.token, err
		}},
		{"service start", func() (string, error) {
			s := &managementServiceContractStub{}
			_, err := NewServicesManagementController(s).Start(ctx, &management.ServiceActionInput{Unit: "ssh.service"})
			return s.token, err
		}},
		{"service stop", func() (string, error) {
			s := &managementServiceContractStub{}
			_, err := NewServicesManagementController(s).Stop(ctx, &management.ServiceActionInput{Unit: "ssh.service"})
			return s.token, err
		}},
		{"service restart", func() (string, error) {
			s := &managementServiceContractStub{}
			_, err := NewServicesManagementController(s).Restart(ctx, &management.ServiceActionInput{Unit: "ssh.service"})
			return s.token, err
		}},
		{"service enable", func() (string, error) {
			s := &managementServiceContractStub{}
			_, err := NewServicesManagementController(s).Enable(ctx, &management.ServiceActionInput{Unit: "ssh.service"})
			return s.token, err
		}},
		{"service disable", func() (string, error) {
			s := &managementServiceContractStub{}
			_, err := NewServicesManagementController(s).Disable(ctx, &management.ServiceActionInput{Unit: "ssh.service"})
			return s.token, err
		}},
		{"sessions list", func() (string, error) {
			s := &managementSessionContractStub{}
			_, err := NewSessionsManagementController(s).List(ctx, &management.ListInput{})
			return s.token, err
		}},
		{"session terminate", func() (string, error) {
			s := &managementSessionContractStub{}
			_, err := NewSessionsManagementController(s).Terminate(ctx, &management.SessionActionInput{Session: "3"})
			return s.token, err
		}},
		{"session lock", func() (string, error) {
			s := &managementSessionContractStub{}
			_, err := NewSessionsManagementController(s).Lock(ctx, &management.SessionActionInput{Session: "3"})
			return s.token, err
		}},
		{"session unlock", func() (string, error) {
			s := &managementSessionContractStub{}
			_, err := NewSessionsManagementController(s).Unlock(ctx, &management.SessionActionInput{Session: "3"})
			return s.token, err
		}},
		{"usb list", func() (string, error) {
			s := &managementUSBContractStub{}
			_, err := NewUSBManagerController(s).List(ctx, &management.ListInput{})
			return s.token, err
		}},
		{"usb allow", func() (string, error) {
			s := &managementUSBContractStub{}
			_, err := NewUSBManagerController(s).Allow(ctx, &management.USBDeviceActionInput{DeviceID: "1"})
			return s.token, err
		}},
		{"usb block", func() (string, error) {
			s := &managementUSBContractStub{}
			_, err := NewUSBManagerController(s).Block(ctx, &management.USBDeviceActionInput{DeviceID: "1"})
			return s.token, err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			token := mustControllerResult(t, tc.call)
			if token != "caller-token" {
				t.Fatalf("forwarded token = %q", token)
			}
		})
	}
}
