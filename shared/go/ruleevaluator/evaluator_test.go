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
