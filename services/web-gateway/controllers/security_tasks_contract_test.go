package controllers

import (
	"context"
	"testing"

	securitydto "lite-nas/services/web-gateway/dto/security_tasks"
	"lite-nas/services/web-gateway/middlewares"
	"lite-nas/shared/authtoken"
	antivirus "lite-nas/shared/contracts/antivirus"
	background "lite-nas/shared/contracts/backgroundtasks"
	integrity "lite-nas/shared/contracts/litenasintegrity"
)

type securityTaskContractStub struct{ token string }

// mustRunControllerAction invokes a controller operation expected to succeed.
func mustRunControllerAction(t *testing.T, action func() error) {
	t.Helper()
	if err := action(); err != nil {
		t.Fatal(err)
	}
}

// mustControllerResult unwraps an expected successful controller result.
func mustControllerResult[T any](t *testing.T, action func() (T, error)) T {
	t.Helper()
	value, err := action()
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func (s *securityTaskContractStub) StartAntivirusScan(_ context.Context, token, _ string) (background.StartResponse, error) {
	s.token = token
	return background.StartResponse{}, nil
}

func (s *securityTaskContractStub) StartIntegrityCheck(_ context.Context, token string) (background.StartResponse, error) {
	s.token = token
	return background.StartResponse{}, nil
}

func (s *securityTaskContractStub) StartIntegrityBaseline(_ context.Context, token string) (background.StartResponse, error) {
	s.token = token
	return background.StartResponse{}, nil
}

func (s *securityTaskContractStub) AntivirusTaskStatus(_ context.Context, token, _ string) (background.StatusCheckResponse, error) {
	s.token = token
	return background.StatusCheckResponse{}, nil
}

func (s *securityTaskContractStub) IntegrityTaskStatus(_ context.Context, token, _ string) (background.StatusCheckResponse, error) {
	s.token = token
	return background.StatusCheckResponse{}, nil
}

func (s *securityTaskContractStub) AntivirusInfo(_ context.Context, token string) (antivirus.InfoResponse, error) {
	s.token = token
	return antivirus.InfoResponse{}, nil
}

func (s *securityTaskContractStub) UpdateAntivirusSignatures(_ context.Context, token string) (antivirus.UpdateSignaturesResponse, error) {
	s.token = token
	return antivirus.UpdateSignaturesResponse{}, nil
}

func (s *securityTaskContractStub) IntegrityInfo(_ context.Context, token string) (integrity.InfoResponse, error) {
	s.token = token
	return integrity.InfoResponse{}, nil
}

func (s *securityTaskContractStub) RetryAntivirusTask(_ context.Context, token, _ string) (background.RetryResponse, error) {
	s.token = token
	return background.RetryResponse{}, nil
}

func (s *securityTaskContractStub) RetryIntegrityTask(_ context.Context, token, _ string) (background.RetryResponse, error) {
	s.token = token
	return background.RetryResponse{}, nil
}

// TestSecurityTaskControllersForwardAccessToken verifies every privileged task endpoint forwards the caller's credential.
func TestSecurityTaskControllersForwardAccessToken(t *testing.T) {
	ctx := middlewares.NewAuthenticatedContext(context.Background(), "caller-token", authtoken.AccessClaims{})
	cases := []struct {
		name string
		call func(SecurityTasksController) error
	}{
		{"start antivirus", func(c SecurityTasksController) error {
			_, err := c.StartAntivirusScan(ctx, &securitydto.StartAntivirusInput{})
			return err
		}},
		{"start integrity", func(c SecurityTasksController) error {
			_, err := c.StartIntegrityCheck(ctx, &securitydto.EmptyInput{})
			return err
		}},
		{"start baseline", func(c SecurityTasksController) error {
			_, err := c.StartIntegrityBaseline(ctx, &securitydto.EmptyInput{})
			return err
		}},
		{"antivirus status", func(c SecurityTasksController) error {
			_, err := c.AntivirusTaskStatus(ctx, &securitydto.TaskIDInput{})
			return err
		}},
		{"integrity status", func(c SecurityTasksController) error {
			_, err := c.IntegrityTaskStatus(ctx, &securitydto.TaskIDInput{})
			return err
		}},
		{"antivirus info", func(c SecurityTasksController) error {
			_, err := c.AntivirusInfo(ctx, &securitydto.EmptyInput{})
			return err
		}},
		{"update signatures", func(c SecurityTasksController) error {
			_, err := c.UpdateAntivirusSignatures(ctx, &securitydto.EmptyInput{})
			return err
		}},
		{"integrity info", func(c SecurityTasksController) error {
			_, err := c.IntegrityInfo(ctx, &securitydto.EmptyInput{})
			return err
		}},
		{"retry antivirus", func(c SecurityTasksController) error {
			_, err := c.RetryAntivirusTask(ctx, &securitydto.TaskIDInput{})
			return err
		}},
		{"retry integrity", func(c SecurityTasksController) error {
			_, err := c.RetryIntegrityTask(ctx, &securitydto.TaskIDInput{})
			return err
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stub := &securityTaskContractStub{}
			mustRunControllerAction(t, func() error { return tc.call(NewSecurityTasksController(stub)) })
			if stub.token != "caller-token" {
				t.Fatalf("forwarded token = %q", stub.token)
			}
		})
	}
}
