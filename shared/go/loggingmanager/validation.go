package loggingmanager

import (
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"

	"lite-nas/shared/loggingmanager/dto"
)

var loggingManagerEventIDPattern = regexp.MustCompile(`^[A-Za-z0-9_]{1,10}_[0-9]{1,8}$`)

// InputValidator validates logging-manager core input DTOs.
type InputValidator interface {
	// Struct validates one DTO against registered schema rules.
	Struct(value any) error
}

// NewInputValidator creates the default validator for logging-manager core
// inputs, including custom filter and event-id validations.
func NewInputValidator() (InputValidator, error) {
	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := validate.RegisterValidation("loggingmanager_event_id", validateLoggingManagerEventID); err != nil {
		return nil, err
	}
	validate.RegisterStructValidation(validateFilterStruct, dto.Filter{})

	return validate, nil
}

func validateLoggingManagerEventID(fieldLevel validator.FieldLevel) bool {
	value := fieldLevel.Field().String()
	return loggingManagerEventIDPattern.MatchString(value)
}

func validateFilterStruct(structLevel validator.StructLevel) {
	filter, ok := structLevel.Current().Interface().(dto.Filter)
	if !ok {
		return
	}

	if !validateFilterValuesCount(structLevel, filter) {
		return
	}

	validateFilterCreatedAtValues(structLevel, filter)
}

func validateFilterValuesCount(structLevel validator.StructLevel, filter dto.Filter) bool {
	valuesCount := len(filter.Values)
	errorTag, isValid := validateConditionValuesCount(filter.Condition, valuesCount)
	if !isValid {
		structLevel.ReportError(filter.Values, "values", "Values", errorTag, "")
		return false
	}

	return true
}

func validateConditionValuesCount(condition dto.FilterCondition, valuesCount int) (string, bool) {
	switch condition {
	case dto.FilterConditionEQ:
		return "eq_values_count", valuesCount == 1
	case dto.FilterConditionIN:
		return "in_values_count", valuesCount >= 1
	case dto.FilterConditionBetween:
		return "between_values_count", valuesCount == 2
	default:
		return "", true
	}
}

func validateFilterCreatedAtValues(structLevel validator.StructLevel, filter dto.Filter) {
	if filter.Key != dto.FilterKeyCreatedAt {
		return
	}

	for _, value := range filter.Values {
		if _, err := time.Parse(time.RFC3339, value); err != nil {
			structLevel.ReportError(filter.Values, "values", "Values", "created_at_rfc3339", "")
			return
		}
	}
}
