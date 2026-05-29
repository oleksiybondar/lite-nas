package ruleevaluator_test

import (
	"testing"

	"lite-nas/shared/ruleevaluator"
)

func TestExtractValueByPathReturnsNestedValue(t *testing.T) {
	t.Parallel()

	value, found := ruleevaluator.ExtractValueByPath(
		map[string]any{
			"snapshot": map[string]any{
				"cpu": map[string]any{
					"totalUsagePct": 91.5,
				},
			},
		},
		"snapshot.cpu.totalUsagePct",
	)

	if !found {
		t.Fatal("found = false, want true")
	}

	if value != 91.5 {
		t.Fatalf("value = %v, want 91.5", value)
	}
}

func TestExtractValuesByPathExpandsArrayTraversal(t *testing.T) {
	t.Parallel()

	values := mustExtractValuesByPath(t, buildPoolHealthPayload(), "snapshot.Pools[].Health")
	assertExtractedValueCount(t, values, 2)
	assertExtractedValue(t, values[0], "snapshot.Pools[0].Health", []int{0}, "ONLINE")
	assertExtractedValue(t, values[1], "snapshot.Pools[1].Health", []int{1}, "DEGRADED")
}

func TestExtractValuesByPathExpandsNestedArrayTraversal(t *testing.T) {
	t.Parallel()

	values := mustExtractValuesByPath(t, buildNestedArrayPayload(), "a.b[].c[].x")
	assertExtractedValueCount(t, values, 3)
	assertExtractedValue(t, values[2], "a.b[1].c[0].x", []int{1, 0}, 20)
}

func buildPoolHealthPayload() map[string]any {
	return map[string]any{
		"snapshot": map[string]any{
			"Pools": []any{
				map[string]any{"Health": "ONLINE"},
				map[string]any{"Health": "DEGRADED"},
			},
		},
	}
}

func buildNestedArrayPayload() map[string]any {
	return map[string]any{
		"a": map[string]any{
			"b": []any{
				map[string]any{
					"c": []any{
						map[string]any{"x": 10},
						map[string]any{"x": 11},
					},
				},
				map[string]any{
					"c": []any{
						map[string]any{"x": 20},
					},
				},
			},
		},
	}
}

func mustExtractValuesByPath(t *testing.T, payload map[string]any, fieldPath string) []ruleevaluator.ExtractedValue {
	t.Helper()

	values, found := ruleevaluator.ExtractValuesByPath(payload, fieldPath)
	if !found {
		t.Fatal("found = false, want true")
	}

	return values
}

func assertExtractedValueCount(t *testing.T, values []ruleevaluator.ExtractedValue, want int) {
	t.Helper()

	if len(values) != want {
		t.Fatalf("len(values) = %d, want %d", len(values), want)
	}
}

func assertExtractedValue(
	t *testing.T,
	value ruleevaluator.ExtractedValue,
	wantFieldPath string,
	wantIndexes []int,
	wantValue any,
) {
	t.Helper()

	if value.FieldPath != wantFieldPath {
		t.Fatalf("value.FieldPath = %q, want %q", value.FieldPath, wantFieldPath)
	}

	if value.Value != wantValue {
		t.Fatalf("value.Value = %v, want %v", value.Value, wantValue)
	}

	assertIndexes(t, value.Indexes, wantIndexes)
}

func assertIndexes(t *testing.T, got []int, want []int) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("len(indexes) = %d, want %d; got=%v", len(got), len(want), got)
	}

	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("indexes[%d] = %d, want %d; got=%v", index, got[index], want[index], got)
		}
	}
}

func TestExtractValueByPathReturnsFalseForUnknownPath(t *testing.T) {
	t.Parallel()

	_, found := ruleevaluator.ExtractValueByPath(
		map[string]any{
			"snapshot": map[string]any{
				"cpu": map[string]any{},
			},
		},
		"snapshot.cpu.totalUsagePct",
	)

	if found {
		t.Fatal("found = true, want false")
	}
}

func TestEvaluateConditionSupportsAllConfiguredConditions(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		extracted any
		condition string
		ruleValue any
		want      bool
	}{
		{name: "greater than", extracted: 91.0, condition: ">", ruleValue: 90.0, want: true},
		{name: "greater than or equal", extracted: 90.0, condition: ">=", ruleValue: 90.0, want: true},
		{name: "equal numeric", extracted: 90, condition: "==", ruleValue: 90.0, want: true},
		{name: "less than or equal", extracted: 90.0, condition: "<=", ruleValue: 90.0, want: true},
		{name: "less than", extracted: 89.0, condition: "<", ruleValue: 90.0, want: true},
		{name: "in", extracted: "warning", condition: "in", ruleValue: []string{"info", "warning", "error"}, want: true},
		{name: "not equal", extracted: "critical", condition: "!=", ruleValue: "warning", want: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := ruleevaluator.EvaluateCondition(testCase.extracted, testCase.condition, testCase.ruleValue)
			if got != testCase.want {
				t.Fatalf("EvaluateCondition() = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestEvaluateConditionRejectsInvalidTypeCombinations(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		extracted any
		condition string
		ruleValue any
	}{
		{name: "greater than with string extracted", extracted: "91", condition: ">", ruleValue: 90},
		{name: "in with non-array rule value", extracted: "warning", condition: "in", ruleValue: "warning"},
		{name: "equal with numeric string rule value", extracted: 90, condition: "==", ruleValue: "90"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if ruleevaluator.EvaluateCondition(testCase.extracted, testCase.condition, testCase.ruleValue) {
				t.Fatal("EvaluateCondition() = true, want false")
			}
		})
	}
}

func TestEvaluateRuleUsesFieldExtractionAndConditionEvaluation(t *testing.T) {
	t.Parallel()

	data := map[string]any{
		"snapshot": map[string]any{
			"mem": map[string]any{
				"usedPct": 90.0,
			},
		},
	}

	rule := ruleevaluator.Rule{
		Field:     "snapshot.mem.usedPct",
		Condition: ">=",
		Values:    90,
	}

	if !ruleevaluator.EvaluateRule(data, rule) {
		t.Fatal("EvaluateRule() = false, want true")
	}
}

func TestEvaluateRuleReturnsFalseWhenPathMissing(t *testing.T) {
	t.Parallel()

	data := map[string]any{
		"snapshot": map[string]any{},
	}

	rule := ruleevaluator.Rule{
		Field:     "snapshot.mem.usedPct",
		Condition: ">=",
		Values:    90,
	}

	if ruleevaluator.EvaluateRule(data, rule) {
		t.Fatal("EvaluateRule() = true, want false")
	}
}
