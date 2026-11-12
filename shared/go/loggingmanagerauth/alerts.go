package loggingmanagerauth

import loggingmanagercontract "lite-nas/shared/contracts/loggingmanager"

// InjectAlertAccessToken adds a service token to supported alert mutation contracts.
// Other payloads pass through unchanged.
func InjectAlertAccessToken(payload any, accessToken string) any {
	switch typed := payload.(type) {
	case loggingmanagercontract.AlertPayload:
		typed.AccessToken = accessToken
		return typed
	case loggingmanagercontract.AlertOccurrencePayload:
		typed.AccessToken = accessToken
		return typed
	case loggingmanagercontract.GetAlertInput:
		typed.AccessToken = accessToken
		return typed
	case loggingmanagercontract.UpdateAlertStateInput:
		typed.AccessToken = accessToken
		return typed
	default:
		return payload
	}
}
