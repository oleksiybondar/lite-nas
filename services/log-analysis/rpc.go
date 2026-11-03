package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/go-playground/validator/v10"

	authcontract "lite-nas/shared/contracts/auth"
	contract "lite-nas/shared/contracts/loganalysis"
	rbaccontract "lite-nas/shared/contracts/rbac"
	"lite-nas/shared/messaging"
	"lite-nas/shared/roleauth"
)

// listRules handles a validated, authorized read of the immutable startup rules.
func listRules(ctx context.Context, envelope messaging.Envelope, spec rulesFile, client messaging.Client) (any, error) {
	if len(envelope.Payload) == 0 || len(envelope.Payload) > 16*1024 {
		return nil, errors.New("invalid log analysis RPC payload size")
	}
	decoder := json.NewDecoder(bytes.NewReader(envelope.Payload))
	decoder.DisallowUnknownFields()
	var request contract.RulesListRequest
	if err := decoder.Decode(&request); err != nil {
		return nil, err
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return nil, errors.New("trailing log analysis RPC JSON")
	}
	if err := validator.New().Struct(request); err != nil {
		return nil, err
	}
	var validated authcontract.ValidateAccessTokenResponse
	if err := client.Request(ctx, authcontract.ValidateAccessTokenRPCSubject, authcontract.ValidateAccessTokenRequest{AccessToken: request.AccessToken}, &validated); err != nil {
		return nil, err
	}
	if !validated.Valid || validated.Status != authcontract.StatusAuthenticated || strings.TrimSpace(validated.Username) == "" {
		return nil, errors.New("security operator authorization required")
	}
	var roles rbaccontract.GetSubjectRolesResponse
	if err := client.Request(ctx, rbaccontract.GetSubjectRolesRPCSubject, rbaccontract.GetSubjectRolesRequest{Username: validated.Username}, &roles); err != nil {
		return nil, err
	}
	if !roleauth.MatchesRequirement(roles.Groups, roleauth.RequirementSecurity) {
		return nil, errors.New("security operator authorization required")
	}
	instances := make(map[string]instanceSpec, len(spec.Instances))
	for _, instance := range spec.Instances {
		instances[instance.ID] = instance
	}
	response := contract.LogAnalysisRulesResponse{Rules: make([]contract.LogAnalysisRule, 0, len(spec.Rules))}
	for _, rule := range spec.Rules {
		instance := instances[rule.Instance]
		response.Rules = append(response.Rules, contract.LogAnalysisRule{
			ID: rule.ID, Name: rule.Name, Explanation: rule.Explanation,
			PossibleReactions: append([]string(nil), rule.PossibleReactions...),
			Enabled:           rule.Enabled, Source: rule.Source, Category: rule.Category,
			Severity: rule.Severity, Instance: rule.Instance, Threshold: instance.Threshold,
			Window: instance.Window, EmailOnCritical: rule.EmailOnCritical, Lockdown: rule.Lockdown,
		})
	}
	return response, nil
}
