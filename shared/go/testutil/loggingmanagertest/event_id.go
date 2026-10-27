package loggingmanagertest

import (
	"testing"

	sharedloggingmanager "lite-nas/shared/loggingmanager"
	loggingmanagerdto "lite-nas/shared/loggingmanager/dto"
)

// MustValidateEventID fails the calling test if an ID violates the manager's
// event-create contract. It is shared by service modules that generate IDs.
func MustValidateEventID(t testing.TB, eventID string) {
	t.Helper()
	validate, err := sharedloggingmanager.NewInputValidator()
	if err != nil {
		t.Fatalf("create logging-manager validator: %v", err)
	}
	if err := validate.Struct(loggingmanagerdto.CreateEventInput{EventID: eventID, Category: "test"}); err != nil {
		t.Fatalf("event ID %q violates logging-manager contract: %v", eventID, err)
	}
}
