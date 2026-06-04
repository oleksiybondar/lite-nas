package roleauth

import "strings"

// Requirement identifies one shared coarse-grained authorization rule used across LiteNAS services.
type Requirement string

const (
	// RequirementOperator grants operator-level monitoring and system-management access.
	RequirementOperator Requirement = "operator"
	// RequirementAdministrator grants administrator-equivalent access.
	RequirementAdministrator Requirement = "administrator"
	// RequirementSecurity grants security-management access.
	RequirementSecurity Requirement = "security"
)

const (
	// RoleOperator is the canonical operator group issued in JWT role claims.
	RoleOperator = "lite-nas-operator"
	// RoleSecurity is the canonical security group issued in JWT role claims.
	RoleSecurity = "lite-nas-security"
	// RoleAdmin is the administrator-equivalent host role.
	RoleAdmin = "admin"
	// RoleSudo is the elevated host role equivalent to administrator access.
	RoleSudo = "sudo"
)

var (
	operatorRoles      = []string{RoleOperator}
	administratorRoles = []string{RoleAdmin, RoleSudo}
	securityRoles      = []string{RoleSecurity}
)

// OperatorRoles returns the canonical target roles that satisfy operator access directly.
func OperatorRoles() []string {
	return cloneRoles(operatorRoles)
}

// AdministratorRoles returns the canonical elevated roles that satisfy administrator access.
func AdministratorRoles() []string {
	return cloneRoles(administratorRoles)
}

// SecurityRoles returns the canonical target roles that satisfy security access directly.
func SecurityRoles() []string {
	return cloneRoles(securityRoles)
}

// AllowedRoles returns the full accepted role set for one shared requirement, including administrator overrides where applicable.
func AllowedRoles(requirement Requirement) []string {
	switch requirement {
	case RequirementOperator:
		return appendRoles(operatorRoles, administratorRoles)
	case RequirementAdministrator:
		return cloneRoles(administratorRoles)
	case RequirementSecurity:
		return appendRoles(securityRoles, administratorRoles)
	default:
		return nil
	}
}

// MatchesRequirement reports whether the provided subject roles satisfy the shared requirement.
func MatchesRequirement(subjectRoles []string, requirement Requirement) bool {
	return HasAnyRole(subjectRoles, AllowedRoles(requirement))
}

// HasAnyRole reports whether the subject holds any of the accepted roles using case-insensitive trimmed matching.
func HasAnyRole(subjectRoles []string, acceptedRoles []string) bool {
	if len(acceptedRoles) == 0 {
		return false
	}

	roleSet := buildNormalizedRoleSet(subjectRoles)
	for _, role := range acceptedRoles {
		key := NormalizeRole(role)
		if key == "" {
			continue
		}
		if _, ok := roleSet[key]; ok {
			return true
		}
	}

	return false
}

// NormalizeRole normalizes one role string for role-set comparisons.
func NormalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func buildNormalizedRoleSet(roles []string) map[string]struct{} {
	roleSet := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		key := NormalizeRole(role)
		if key == "" {
			continue
		}
		roleSet[key] = struct{}{}
	}
	return roleSet
}

func cloneRoles(roles []string) []string {
	if len(roles) == 0 {
		return nil
	}

	clone := make([]string, len(roles))
	copy(clone, roles)
	return clone
}

func appendRoles(left []string, right []string) []string {
	combined := make([]string, 0, len(left)+len(right))
	combined = append(combined, left...)
	combined = append(combined, right...)
	return combined
}
