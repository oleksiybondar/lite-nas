package main

import (
	"encoding/hex"
	"strings"

	authcontract "lite-nas/shared/contracts/auth"
	"lite-nas/shared/messaging"
)

// handleCaptionGrantIssueRPC verifies the initiating user and persists one grant.
func handleCaptionGrantIssueRPC(runtimeDeps authRuntime, envelope messaging.Envelope) (authcontract.CaptionGrantIssueResponse, error) {
	var request authcontract.CaptionGrantIssueRequest
	if !decodeRPCRequest(envelope, &request, runtimeDeps.Validator) || runtimeDeps.CaptionGrants == nil {
		return authcontract.CaptionGrantIssueResponse{}, nil
	}
	claims, valid := verifyAccessToken(runtimeDeps, request.UserAccessToken)
	if !valid || strings.TrimSpace(claims.Login) == "" {
		return authcontract.CaptionGrantIssueResponse{}, nil
	}
	seen := make(map[string]bool, len(request.PathHashes))
	for _, hash := range request.PathHashes {
		if _, err := hex.DecodeString(hash); err != nil || seen[hash] {
			return authcontract.CaptionGrantIssueResponse{}, nil
		}
		seen[hash] = true
	}
	grant, expiresAt, err := runtimeDeps.CaptionGrants.issue(request.JobID, claims.Login, request.PathHashes)
	if err != nil {
		return authcontract.CaptionGrantIssueResponse{}, err
	}
	return authcontract.CaptionGrantIssueResponse{Grant: grant, ExpiresAt: expiresAt}, nil
}

// handleCaptionGrantValidateRPC checks an exact File Service path online.
func handleCaptionGrantValidateRPC(runtimeDeps authRuntime, envelope messaging.Envelope) (authcontract.CaptionGrantValidateResponse, error) {
	var request authcontract.CaptionGrantValidateRequest
	if !decodeRPCRequest(envelope, &request, runtimeDeps.Validator) || runtimeDeps.CaptionGrants == nil {
		return authcontract.CaptionGrantValidateResponse{}, nil
	}
	username := runtimeDeps.CaptionGrants.validate(request.Grant, request.Path)
	return authcontract.CaptionGrantValidateResponse{Valid: username != "", Username: username}, nil
}

// handleCaptionGrantRevokeRPC removes the credential after a job ends.
func handleCaptionGrantRevokeRPC(runtimeDeps authRuntime, envelope messaging.Envelope) (authcontract.CaptionGrantRevokeResponse, error) {
	var request authcontract.CaptionGrantRevokeRequest
	if !decodeRPCRequest(envelope, &request, runtimeDeps.Validator) || runtimeDeps.CaptionGrants == nil {
		return authcontract.CaptionGrantRevokeResponse{}, nil
	}
	revoked, err := runtimeDeps.CaptionGrants.revoke(request.Grant)
	return authcontract.CaptionGrantRevokeResponse{Revoked: revoked}, err
}
