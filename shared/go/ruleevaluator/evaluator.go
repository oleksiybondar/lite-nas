package ruleevaluator

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Rule describes one condition check against an extracted payload field.
type Rule struct {
	Field     string
	Condition string
	Values    any
}

// ExtractedValue stores one value resolved from a rule field path, including
// any array indexes traversed and the fully indexed field path.
type ExtractedValue struct {
	FieldPath string
	Indexes   []int
	Value     any
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
	matches, found := ExtractValuesByPath(data, fieldPath)
	if !found || len(matches) != 1 {
		return nil, false
	}

	return matches[0].Value, true
}

// ExtractValuesByPath resolves a dot-separated field path and expands any
// array traversal segment written as "name[]", returning one extracted value
// per resolved path such as "snapshot.Pools[0].Health".
func ExtractValuesByPath(data map[string]any, fieldPath string) ([]ExtractedValue, bool) {
	pathParts := splitFieldPath(fieldPath)
	if len(pathParts) == 0 {
		return nil, false
	}

	currentMatches, ok := expandPathMatches(pathParts, []pathMatch{{value: data}})
	if !ok {
		return nil, false
	}

	return buildExtractedValues(currentMatches), true
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

// pathPart represents one parsed field path segment and whether it expands an
// array traversal.
type pathPart struct {
	name        string
	isArrayWalk bool
}

// pathMatch stores one in-progress traversal branch while expanding array
// segments.
type pathMatch struct {
	fieldParts []string
	indexes    []int
	value      any
}

// splitFieldPath parses a rule field path into traversal segments.
func splitFieldPath(fieldPath string) []pathPart {
	trimmedPath := strings.TrimSpace(fieldPath)
	if trimmedPath == "" {
		return nil
	}

	rawParts := strings.Split(trimmedPath, ".")
	parts := make([]pathPart, 0, len(rawParts))

	for _, rawPart := range rawParts {
		segment := strings.TrimSpace(rawPart)
		if segment == "" {
			return nil
		}

		parts = append(parts, pathPart{
			name:        strings.TrimSuffix(segment, "[]"),
			isArrayWalk: strings.HasSuffix(segment, "[]"),
		})
	}

	return parts
}

// resolvePathPart advances one traversal branch by one field path segment.
func resolvePathPart(currentMatch pathMatch, part pathPart) []pathMatch {
	node, ok := currentMatch.value.(map[string]any)
	if !ok {
		return nil
	}

	nextValue, exists := node[part.name]
	if !exists {
		return nil
	}

	if !part.isArrayWalk {
		return []pathMatch{{
			fieldParts: appendPathPart(currentMatch.fieldParts, part.name),
			indexes:    append([]int(nil), currentMatch.indexes...),
			value:      nextValue,
		}}
	}

	elements, ok := toAnySlice(nextValue)
	if !ok {
		return nil
	}

	matches := make([]pathMatch, 0, len(elements))
	for index, element := range elements {
		matches = append(matches, pathMatch{
			fieldParts: appendPathPart(currentMatch.fieldParts, fmt.Sprintf("%s[%d]", part.name, index)),
			indexes:    appendIndex(currentMatch.indexes, index),
			value:      element,
		})
	}

	return matches
}

// expandPathMatches walks parsed path parts across the current traversal
// branches, expanding array segments when needed.
func expandPathMatches(pathParts []pathPart, currentMatches []pathMatch) ([]pathMatch, bool) {
	for _, pathPart := range pathParts {
		nextMatches := resolvePathMatches(currentMatches, pathPart)
		if len(nextMatches) == 0 {
			return nil, false
		}

		currentMatches = nextMatches
	}

	return currentMatches, true
}

// resolvePathMatches applies one path segment to all in-progress traversal
// branches.
func resolvePathMatches(currentMatches []pathMatch, part pathPart) []pathMatch {
	nextMatches := make([]pathMatch, 0, len(currentMatches))
	for _, currentMatch := range currentMatches {
		nextMatches = append(nextMatches, resolvePathPart(currentMatch, part)...)
	}

	return nextMatches
}

// buildExtractedValues converts completed traversal branches into exported
// extracted-value records.
func buildExtractedValues(matches []pathMatch) []ExtractedValue {
	result := make([]ExtractedValue, 0, len(matches))
	for _, currentMatch := range matches {
		result = append(result, ExtractedValue{
			FieldPath: strings.Join(currentMatch.fieldParts, "."),
			Indexes:   append([]int(nil), currentMatch.indexes...),
			Value:     currentMatch.value,
		})
	}

	return result
}

// appendPathPart returns a copied path-parts slice with one segment appended.
func appendPathPart(parts []string, part string) []string {
	appended := make([]string, 0, len(parts)+1)
	appended = append(appended, parts...)
	appended = append(appended, part)
	return appended
}

// appendIndex returns a copied indexes slice with one index appended.
func appendIndex(indexes []int, index int) []int {
	appended := make([]int, 0, len(indexes)+1)
	appended = append(appended, indexes...)
	appended = append(appended, index)
	return appended
}

// FormatIndexQualifiers converts rule traversal indexes into cache key
// qualifiers using decimal string segments.
func FormatIndexQualifiers(indexes []int) []string {
	qualifiers := make([]string, 0, len(indexes))
	for _, index := range indexes {
		qualifiers = append(qualifiers, strconv.Itoa(index))
	}

	return qualifiers
}
