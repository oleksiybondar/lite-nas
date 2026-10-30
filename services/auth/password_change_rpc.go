package main

import (
	"context"

	"lite-nas/services/auth/pamauth"
	authcontract "lite-nas/shared/contracts/auth"
	"lite-nas/shared/messaging"
)

// handlePasswordChangeStartRPC begins an on-demand PAM change for the JWT user.
func handlePasswordChangeStartRPC(ctx context.Context, deps authRuntime, envelope messaging.Envelope) (authcontract.LoginResponse, error) {
	var request authcontract.PasswordChangeStartRequest
	if !decodeRPCRequest(envelope, &request, deps.Validator) || deps.Conversations == nil || lockdownActive(deps) {
		return authcontract.LoginResponse{Status: authcontract.StatusDenied}, nil
	}
	claims, valid := verifyAccessToken(deps, request.AccessToken)
	if !valid || claims.Login == "" {
		return authcontract.LoginResponse{Status: authcontract.StatusDenied}, nil
	}
	event, flow, ok := deps.Conversations.start(ctx, authcontract.LoginRequest{Username: claims.Login}, true)
	if !ok {
		return authcontract.LoginResponse{Status: authcontract.StatusDenied}, nil
	}
	return loginResponseForConversation(deps, event, flow), nil
}

// handlePasswordChangeReplyRPC forwards one browser answer to the owning PAM worker.
func handlePasswordChangeReplyRPC(ctx context.Context, deps authRuntime, envelope messaging.Envelope) (authcontract.LoginResponse, error) {
	var request authcontract.PasswordChangeReplyRequest
	if !decodeRPCRequest(envelope, &request, deps.Validator) || deps.Conversations == nil || lockdownActive(deps) {
		return authcontract.LoginResponse{Status: authcontract.StatusDenied}, nil
	}
	event, flow, ok := deps.Conversations.step(ctx, request.ConversationID, request.Sequence, request.Answer)
	if !ok {
		return authcontract.LoginResponse{Status: authcontract.StatusDenied}, nil
	}
	return loginResponseForConversation(deps, event, flow), nil
}

// loginResponseForConversation issues a session only after PAM has completed.
func loginResponseForConversation(deps authRuntime, event passwordEvent, flow *passwordFlow) authcontract.LoginResponse {
	if event.prompt != nil {
		return authcontract.LoginResponse{
			Status:            authcontract.StatusPasswordChangeRequired,
			Username:          flow.login.Username,
			CanChangePassword: true,
			Conversation:      &authcontract.PasswordConversation{ID: flow.id, Sequence: flow.sequence.Load(), Prompt: event.prompt.Text, Secret: event.prompt.Secret, ExpiresAt: flow.expiresAt, Messages: authMessagesFromPAM(event.prompt.Messages)},
		}
	}
	result := event.result
	if result.Code != pamauth.OutcomeAuthenticated {
		return loginResponseFromPAM(result)
	}
	if flow.onDemand {
		return loginResponseFromPAM(result)
	}
	if deps.Lockdown != nil {
		deps.Lockdown.mu.RLock()
		defer deps.Lockdown.mu.RUnlock()
		if deps.Lockdown.state.Lockdown {
			return authcontract.LoginResponse{Status: authcontract.StatusDenied}
		}
	}
	return issueLoginSession(deps, flow.login, result)
}

// lockdownActive reports the current emergency auth policy before admitting work.
func lockdownActive(deps authRuntime) bool {
	if deps.Lockdown == nil {
		return false
	}
	deps.Lockdown.mu.RLock()
	defer deps.Lockdown.mu.RUnlock()
	return deps.Lockdown.state.Lockdown
}
