package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	authcontract "lite-nas/shared/contracts/auth"
	contract "lite-nas/shared/contracts/loganalysis"
	rbaccontract "lite-nas/shared/contracts/rbac"
	"lite-nas/shared/messaging"
)

// rulesClient answers live authorization calls for the catalog handler tests.
type rulesClient struct{ security bool }

func (*rulesClient) Publish(context.Context, string, any) error { return nil }
func (*rulesClient) Drain() error                               { return nil }
func (*rulesClient) Close()                                     {}

// Request supplies the two existing authorization contracts used by the handler.
func (c *rulesClient) Request(_ context.Context, subject string, _ any, response any) error {
	switch subject {
	case authcontract.ValidateAccessTokenRPCSubject:
		*response.(*authcontract.ValidateAccessTokenResponse) = authcontract.ValidateAccessTokenResponse{Valid: true, Status: authcontract.StatusAuthenticated, Username: "alice"}
	case rbaccontract.GetSubjectRolesRPCSubject:
		if c.security {
			response.(*rbaccontract.GetSubjectRolesResponse).Groups = []string{"lite-nas-security"}
		}
	default:
		return errors.New("unexpected subject")
	}
	return nil
}

// mustRulesEnvelope builds a valid request without repeating payload setup.
func mustRulesEnvelope(t *testing.T) messaging.Envelope {
	t.Helper()
	payload, err := json.Marshal(contract.RulesListRequest{AccessToken: "token"})
	if err != nil {
		t.Fatal(err)
	}
	return messaging.Envelope{Payload: payload}
}

// Requirements: security-monitoring/FR-001.
// TestListRulesReturnsOperatorMetadata verifies a configured rule is presented with guidance.
func TestListRulesReturnsOperatorMetadata(t *testing.T) {
	spec, err := loadRulesFile("../../configs/etc/lite-nas/log-analysis/rules.json")
	if err != nil {
		t.Fatal(err)
	}
	result, err := listRules(context.Background(), mustRulesEnvelope(t), spec, &rulesClient{security: true})
	if err != nil {
		t.Fatal(err)
	}
	rules := result.(contract.LogAnalysisRulesResponse).Rules
	if len(rules) != len(spec.Rules) || rules[0].Name == "" || rules[0].Explanation == "" || len(rules[0].PossibleReactions) == 0 {
		t.Fatal("catalog omitted configured rule guidance")
	}
}

// Requirements: security-monitoring/FR-001.
// TestListRulesRejectsNonSecurityRole verifies the NATS boundary enforces the security role.
func TestListRulesRejectsNonSecurityRole(t *testing.T) {
	_, err := listRules(context.Background(), mustRulesEnvelope(t), rulesFile{}, &rulesClient{})
	if err == nil {
		t.Fatal("non-security role accepted")
	}
}

// Requirements: security-monitoring/FR-001.
// TestListRulesRejectsMalformedInput verifies unknown fields are rejected before auth calls.
func TestListRulesRejectsMalformedInput(t *testing.T) {
	_, err := listRules(context.Background(), messaging.Envelope{Payload: []byte(`{"access_token":"token","unexpected":true}`)}, rulesFile{}, &rulesClient{security: true})
	if err == nil {
		t.Fatal("unknown request field accepted")
	}
}
