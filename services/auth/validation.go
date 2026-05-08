package main

import "github.com/go-playground/validator/v10"

// requestValidator validates decoded service-bound request DTOs before handler
// logic observes them.
type requestValidator interface {
	// Struct validates a decoded request DTO using its validation schema tags.
	Struct(value any) error
}

// newRequestValidator creates the default validator used by auth RPC handlers.
func newRequestValidator() requestValidator {
	return validator.New(validator.WithRequiredStructEnabled())
}

// validateRPCRequest applies the injected request validator to a decoded RPC
// DTO.
func validateRPCRequest(validator requestValidator, request any) bool {
	if validator == nil {
		return false
	}

	return validator.Struct(request) == nil
}
