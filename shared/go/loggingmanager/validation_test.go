package loggingmanager

import (
	"testing"

	"lite-nas/shared/loggingmanager/dto"
)

func TestInputValidatorRejectsInvalidEventIDPattern(t *testing.T) {
	t.Parallel()

	validate := mustInputValidator(t)
	err := validate.Struct(dto.SetStateInput{
		EventID: "invalid-id",
		Status:  "active",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestInputValidatorAcceptsCompactTimestampEventIDPattern(t *testing.T) {
	t.Parallel()

	validate := mustInputValidator(t)
	err := validate.Struct(dto.SetStateInput{
		EventID: "t1778675852000000000",
		Status:  "active",
	})
	if err != nil {
		t.Fatalf("validation error = %v", err)
	}
}

func TestInputValidatorRejectsInvalidFilters(t *testing.T) {
	t.Parallel()

	validate := mustInputValidator(t)
	cases := []struct {
		name   string
		filter dto.Filter
	}{
		{
			name: "between filter with one value",
			filter: dto.Filter{
				Key:       dto.FilterKeyCreatedAt,
				Condition: dto.FilterConditionBetween,
				Values:    []string{"2026-05-12T10:00:00Z"},
			},
		},
		{
			name: "created_at non rfc3339",
			filter: dto.Filter{
				Key:       dto.FilterKeyCreatedAt,
				Condition: dto.FilterConditionEQ,
				Values:    []string{"2026-05-12"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := validate.Struct(dto.ListEventsInput{
				Page:    1,
				Filters: []dto.Filter{tc.filter},
			})
			if err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestInputValidatorAcceptsValidListEventsInput(t *testing.T) {
	t.Parallel()

	validate := mustInputValidator(t)
	err := validate.Struct(dto.ListEventsInput{
		Page:     1,
		PageSize: 25,
		Filters: []dto.Filter{
			{
				Key:       dto.FilterKeyCategory,
				Condition: dto.FilterConditionIN,
				Values:    []string{"system", "security"},
			},
		},
	})
	if err != nil {
		t.Fatalf("validation error = %v", err)
	}
}

func mustInputValidator(t *testing.T) InputValidator {
	t.Helper()
	validate, err := NewInputValidator()
	if err != nil {
		t.Fatalf("NewInputValidator() error = %v", err)
	}
	return validate
}
