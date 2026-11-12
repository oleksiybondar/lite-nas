package services

import (
	"context"
	"errors"
	"testing"

	servicecontract "lite-nas/shared/contracts/servicesmanager"
	sessioncontract "lite-nas/shared/contracts/sessionmanager"
)

// Requirements: web-gateway/FR-003, services-manager/FR-001, services-manager/FR-002.
func TestServicesManagementRPCSubjects(t *testing.T) {
	cases := []struct{ action, subject string }{{"start", servicecontract.StartSubject}, {"stop", servicecontract.StopSubject}, {"restart", servicecontract.RestartSubject}, {"enable", servicecontract.EnableSubject}, {"disable", servicecontract.DisableSubject}}
	for _, tc := range cases {
		t.Run(tc.action, func(t *testing.T) {
			client := &alertsClientStub{requestFunc: func(_ context.Context, _ string, request, reply any) error {
				if request.(servicecontract.ActionRequest).Unit != "ssh.service" {
					t.Fatal("wrong unit")
				}
				reply.(*servicecontract.ActionResponse).Completed = true
				return nil
			}}
			if err := NewServicesManagementService(client).Action(context.Background(), "token", "ssh.service", tc.action); err != nil {
				t.Fatal(err)
			}
			if client.subject != tc.subject {
				t.Fatalf("subject=%q, want %q", client.subject, tc.subject)
			}
		})
	}
}

// Requirements: web-gateway/FR-003, session-manager/FR-001, session-manager/FR-002.
func TestSessionsManagementRPCSubjects(t *testing.T) {
	cases := []struct{ action, subject string }{{"terminate", sessioncontract.TerminateSubject}, {"lock", sessioncontract.LockSubject}, {"unlock", sessioncontract.UnlockSubject}}
	for _, tc := range cases {
		t.Run(tc.action, func(t *testing.T) {
			client := &alertsClientStub{requestFunc: func(_ context.Context, _ string, request, reply any) error {
				if request.(sessioncontract.ActionRequest).Session != "3" {
					t.Fatal("wrong session")
				}
				reply.(*sessioncontract.ActionResponse).Completed = true
				return nil
			}}
			if err := NewSessionsManagementService(client).Action(context.Background(), "token", "3", tc.action); err != nil {
				t.Fatal(err)
			}
			if client.subject != tc.subject {
				t.Fatalf("subject=%q, want %q", client.subject, tc.subject)
			}
		})
	}
}

// Requirements: web-gateway/FR-003.
func TestManagementListsForwardToken(t *testing.T) {
	t.Run("services", func(t *testing.T) {
		client := &alertsClientStub{requestFunc: func(_ context.Context, _ string, request, reply any) error {
			if request.(servicecontract.ListRequest).AccessToken != "token" {
				t.Fatal("wrong token")
			}
			reply.(*servicecontract.ListResponse).Services = []servicecontract.Service{{Name: "ssh.service"}}
			return nil
		}}
		items, err := NewServicesManagementService(client).List(context.Background(), "token")
		if err != nil {
			t.Fatal(err)
		}
		if len(items) != 1 || client.subject != servicecontract.ListSubject {
			t.Fatal("wrong list result or subject")
		}
	})
	t.Run("sessions", func(t *testing.T) {
		client := &alertsClientStub{requestFunc: func(_ context.Context, _ string, request, reply any) error {
			if request.(sessioncontract.ListRequest).AccessToken != "token" {
				t.Fatal("wrong token")
			}
			reply.(*sessioncontract.ListResponse).Sessions = []sessioncontract.Session{{ID: "3"}}
			return nil
		}}
		items, err := NewSessionsManagementService(client).List(context.Background(), "token")
		if err != nil {
			t.Fatal(err)
		}
		if len(items) != 1 || client.subject != sessioncontract.ListSubject {
			t.Fatal("wrong list result or subject")
		}
	})
}

// Requirements: web-gateway/FR-003.
func TestManagementRejectsUnconfirmedAction(t *testing.T) {
	client := &alertsClientStub{}
	err := NewServicesManagementService(client).Action(context.Background(), "token", "ssh.service", "start")
	if !errors.Is(err, ErrManagementActionFailed) {
		t.Fatalf("error=%v", err)
	}
}
