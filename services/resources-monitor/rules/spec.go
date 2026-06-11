package rules

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strings"

	validation "github.com/go-playground/validator/v10"

	sharedloggingenum "lite-nas/shared/loggingmanager/enum"
)

var (
	errEmptyRulesFiles         = errors.New("rules files are required")
	errRuleValueType           = errors.New("rule values type is invalid for condition")
	errRulePrefixTooLong       = errors.New("rule event_prefix must be at most 8 characters")
	errRuleNumericValuesNeeded = errors.New("numeric condition requires numeric rule values")
)

// Rule defines one threshold rule loaded from a JSON rules file.
type Rule struct {
	Event         string                     `json:"event" validate:"required,min=1,max=128"`
	EventPrefix   string                     `json:"event_prefix" validate:"required,min=1,max=8"`
	Field         string                     `json:"field" validate:"required,min=1,max=256"`
	Condition     string                     `json:"condition" validate:"required,oneof=> >= == <= < in !="`
	Values        any                        `json:"values" validate:"required"`
	Message       string                     `json:"message" validate:"required,min=1,max=256"`
	Description   string                     `json:"description,omitempty" validate:"omitempty,max=512"`
	NormalMessage string                     `json:"normal_message,omitempty" validate:"omitempty,max=256"`
	Category      string                     `json:"category" validate:"required,min=1,max=128"`
	Severity      sharedloggingenum.Severity `json:"severity" validate:"required,oneof=info warning error critical"`
	Priority      int                        `json:"priority" validate:"gte=0,lte=5"`
	Source        string                     `json:"source" validate:"required,min=1,max=128"`
}

// LoadRules loads and validates rules from one or more JSON files.
func LoadRules(files []string) ([]Rule, error) {
	if len(files) == 0 {
		return nil, errEmptyRulesFiles
	}

	validate := validation.New(validation.WithRequiredStructEnabled())
	allRules := make([]Rule, 0, 8)

	for _, file := range files {
		fileRules, err := loadAndValidateRulesFile(file, validate)
		if err != nil {
			return nil, err
		}

		allRules = append(allRules, fileRules...)
	}

	return allRules, nil
}

// loadAndValidateRulesFile decodes and validates all rules from one rules file.
func loadAndValidateRulesFile(file string, validate *validation.Validate) ([]Rule, error) {
	fileRules, err := loadRulesFile(file)
	if err != nil {
		return nil, err
	}

	if err = validateRules(file, fileRules, validate); err != nil {
		return nil, err
	}

	return fileRules, nil
}

// validateRules validates one decoded rules slice from a single file.
func validateRules(file string, fileRules []Rule, validate *validation.Validate) error {
	for index, rule := range fileRules {
		if err := validate.Struct(rule); err != nil {
			return fmt.Errorf("rules file %q entry %d validation failed: %w", file, index, err)
		}

		if err := validateRuleValues(rule); err != nil {
			return fmt.Errorf("rules file %q entry %d is invalid: %w", file, index, err)
		}
	}

	return nil
}

// loadRulesFile reads and decodes one rules JSON file.
func loadRulesFile(file string) ([]Rule, error) {
	rawData, err := os.ReadFile(file) // #nosec G304 -- file path is configured by service config.
	if err != nil {
		return nil, err
	}

	var loadedRules []Rule
	if err = json.Unmarshal(rawData, &loadedRules); err != nil {
		return nil, err
	}

	return loadedRules, nil
}

// validateRuleValues enforces condition-specific rule values contract.
func validateRuleValues(rule Rule) error {
	if len(strings.TrimSpace(rule.EventPrefix)) > 8 {
		return errRulePrefixTooLong
	}

	if rule.Condition == "in" {
		if !isSliceOrArray(rule.Values) {
			return errRuleValueType
		}

		return nil
	}

	if !slices.Contains([]string{">", ">=", "<", "<=", "=="}, rule.Condition) {
		return nil
	}

	if !isNumber(rule.Values) {
		return errRuleNumericValuesNeeded
	}

	return nil
}

// isSliceOrArray reports whether value is a slice or array.
func isSliceOrArray(value any) bool {
	kind := reflect.TypeOf(value)
	if kind == nil {
		return false
	}

	return kind.Kind() == reflect.Slice || kind.Kind() == reflect.Array
}

// isNumber reports whether value is any supported numeric type.
func isNumber(value any) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	default:
		return false
	}
}
