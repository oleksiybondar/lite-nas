package roleauth

import (
	"reflect"
	"testing"
)

func TestAllowedRolesReturnsCanonicalRoleSets(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		requirement Requirement
		want        []string
	}{
		{name: "operator", requirement: RequirementOperator, want: []string{RoleOperator, RoleAdmin, RoleSudo}},
		{name: "administrator", requirement: RequirementAdministrator, want: []string{RoleAdmin, RoleSudo}},
		{name: "security", requirement: RequirementSecurity, want: []string{RoleSecurity, RoleAdmin, RoleSudo}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := AllowedRoles(testCase.requirement)
			if !reflect.DeepEqual(got, testCase.want) {
				t.Fatalf("AllowedRoles() = %#v, want %#v", got, testCase.want)
			}
		})
	}
}

func TestMatchesRequirementUsesAdministratorOverride(t *testing.T) {
	t.Parallel()

	if !MatchesRequirement([]string{" sudo "}, RequirementSecurity) {
		t.Fatal("expected sudo to satisfy security requirement")
	}

	if MatchesRequirement([]string{RoleOperator}, RequirementSecurity) {
		t.Fatal("expected operator role to fail security requirement")
	}
}

func TestHasAnyRoleNormalizesRoleNames(t *testing.T) {
	t.Parallel()

	if !HasAnyRole([]string{" Lite-Nas-Operator "}, []string{RoleOperator}) {
		t.Fatal("expected normalized operator role to match")
	}

	if HasAnyRole([]string{"viewer"}, nil) {
		t.Fatal("expected empty accepted role set to reject")
	}
}
