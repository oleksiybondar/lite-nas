package ruleevaluator

import (
	"reflect"
	"strings"
)

// Rule describes one condition check against an extracted payload field.
type Rule struct {
	Field     string
	Condition string
	Values    any
}

// EvaluateRule extracts a value from input data using rule.Field and evaluates
// it against rule.Condition and rule.Values.
func EvaluateRule(data map[string]any, rule Rule) bool {
	extractedValue, found := ExtractValueByPath(data, rule.Field)
	if !found {
		return false
	}

	return EvaluateCondition(extractedValue, rule.Condition, rule.Values)
}

// EvaluateCondition routes condition evaluation to a dedicated condition
// function.
func EvaluateCondition(extractedValue any, condition string, ruleValue any) bool {
	evaluatorByCondition := map[string]func(any, any) bool{
		">":  EvaluateGreaterThan,
		">=": EvaluateGreaterThanOrEqual,
		"==": EvaluateEqual,
		"<=": EvaluateLessThanOrEqual,
		"<":  EvaluateLessThan,
		"in": EvaluateIn,
		"!=": EvaluateNotEqual,
	}

	evaluator, exists := evaluatorByCondition[condition]
	if !exists {
		return false
	}

	return evaluator(extractedValue, ruleValue)
}

// ExtractValueByPath resolves a dot-separated field path (for example
// "snapshot.cpu.totalUsagePct") from a nested object tree.
func ExtractValueByPath(data map[string]any, fieldPath string) (any, bool) {
	trimmedPath := strings.TrimSpace(fieldPath)
	if trimmedPath == "" {
		return nil, false
	}

	pathParts := strings.Split(trimmedPath, ".")
	current := any(data)

	for _, pathPart := range pathParts {
		node, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}

		next, exists := node[pathPart]
		if !exists {
			return nil, false
		}

		current = next
	}

	return current, true
}

// EvaluateGreaterThan evaluates extractedValue > ruleValue for numeric values.
func EvaluateGreaterThan(extractedValue any, ruleValue any) bool {
	left, ok := toFloat64(extractedValue)
	if !ok {
		return false
	}

	right, ok := toFloat64(ruleValue)
	if !ok {
		return false
	}

	return left > right
}

// EvaluateGreaterThanOrEqual evaluates extractedValue >= ruleValue for numeric
// values.
func EvaluateGreaterThanOrEqual(extractedValue any, ruleValue any) bool {
	left, ok := toFloat64(extractedValue)
	if !ok {
		return false
	}

	right, ok := toFloat64(ruleValue)
	if !ok {
		return false
	}

	return left >= right
}

// EvaluateEqual evaluates extractedValue == ruleValue.
//
// For numeric values, numeric equality is used across numeric types. For all
// other values, deep equality is used.
func EvaluateEqual(extractedValue any, ruleValue any) bool {
	left, leftIsNumber := toFloat64(extractedValue)
	right, rightIsNumber := toFloat64(ruleValue)
	if leftIsNumber && rightIsNumber {
		return left == right
	}

	return reflect.DeepEqual(extractedValue, ruleValue)
}

// EvaluateLessThanOrEqual evaluates extractedValue <= ruleValue for numeric
// values.
func EvaluateLessThanOrEqual(extractedValue any, ruleValue any) bool {
	left, ok := toFloat64(extractedValue)
	if !ok {
		return false
	}

	right, ok := toFloat64(ruleValue)
	if !ok {
		return false
	}

	return left <= right
}

// EvaluateLessThan evaluates extractedValue < ruleValue for numeric values.
func EvaluateLessThan(extractedValue any, ruleValue any) bool {
	left, ok := toFloat64(extractedValue)
	if !ok {
		return false
	}

	right, ok := toFloat64(ruleValue)
	if !ok {
		return false
	}

	return left < right
}

// EvaluateIn evaluates extractedValue membership in ruleValue when ruleValue is
// a slice or array.
func EvaluateIn(extractedValue any, ruleValue any) bool {
	candidates, ok := toAnySlice(ruleValue)
	if !ok {
		return false
	}

	for _, candidate := range candidates {
		if EvaluateEqual(extractedValue, candidate) {
			return true
		}
	}

	return false
}

// EvaluateNotEqual evaluates extractedValue != ruleValue.
func EvaluateNotEqual(extractedValue any, ruleValue any) bool {
	return !EvaluateEqual(extractedValue, ruleValue)
}

// toFloat64 converts supported numeric values to float64.
func toFloat64(value any) (float64, bool) {
	reflectedValue := reflect.ValueOf(value)
	if !reflectedValue.IsValid() {
		return 0, false
	}

	switch reflectedValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(reflectedValue.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(reflectedValue.Uint()), true
	case reflect.Float32, reflect.Float64:
		return reflectedValue.Float(), true
	default:
		return 0, false
	}
}

// toAnySlice converts slices and arrays to []any.
func toAnySlice(value any) ([]any, bool) {
	reflectedValue := reflect.ValueOf(value)
	if !reflectedValue.IsValid() {
		return nil, false
	}

	valueKind := reflectedValue.Kind()
	if valueKind != reflect.Slice && valueKind != reflect.Array {
		return nil, false
	}

	result := make([]any, reflectedValue.Len())
	for index := range reflectedValue.Len() {
		result[index] = reflectedValue.Index(index).Interface()
	}

	return result, true
}
