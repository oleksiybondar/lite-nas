package testcasetest

import (
	"reflect"
	"testing"
)

// FieldCase defines one named field assertion over a loaded fixture value.
type FieldCase[T any] struct {
	Name string
	Got  func(T) any
	Want any
}

// RunFieldCases executes the common "load fixture, select field, compare"
// matrix pattern used across LiteNAS tests.
func RunFieldCases[T any](
	t *testing.T,
	load func(*testing.T) T,
	testCases []FieldCase[T],
) {
	t.Helper()

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			value := load(t)
			if got := testCase.Got(value); !reflect.DeepEqual(got, testCase.Want) {
				t.Fatalf("%s = %#v, want %#v", testCase.Name, got, testCase.Want)
			}
		})
	}
}
