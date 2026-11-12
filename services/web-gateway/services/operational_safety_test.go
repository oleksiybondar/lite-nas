package services

import (
	"context"
	"testing"

	contract "lite-nas/shared/contracts/operationalsafety"
)

// TestOperationalSafetyServiceRoutesOperatorRPCs verifies each gateway call's subject.
func TestOperationalSafetyServiceRoutesOperatorRPCs(t *testing.T) {
	ctx := context.Background()
	for _, test := range []struct {
		name, subject string
		call          func(OperationalSafetyService) error
	}{
		{"rules", contract.RulesListSubject, func(s OperationalSafetyService) error { _, err := s.Rules(ctx, "token"); return err }},
		{"tasks", contract.TasksListSubject, func(s OperationalSafetyService) error { _, err := s.Tasks(ctx, "token"); return err }},
		{"rule state", contract.RuleSetEnabledSubject, func(s OperationalSafetyService) error {
			enabled := true
			_, err := s.SetEnabled(ctx, contract.OperationalSafetyRuleSetEnabledRequest{AccessToken: "token", RuleID: "rule", Enabled: &enabled})
			return err
		}},
		{"force set", contract.TaskForceSetSubject, func(s OperationalSafetyService) error {
			_, err := s.ForceSet(ctx, contract.OperationalSafetyTaskForceSetRequest{AccessToken: "token", TaskID: "task", Level: contract.LevelWarning})
			return err
		}},
		{"force stop", contract.TaskForceStopSubject, func(s OperationalSafetyService) error {
			_, err := s.ForceStop(ctx, contract.OperationalSafetyTaskForceStopRequest{AccessToken: "token", TaskID: "task"})
			return err
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			client := &metricsClientStub{}
			if err := test.call(NewOperationalSafetyService(client)); err != nil {
				t.Fatal(err)
			}
			assertMetricsSubject(t, client.subject, test.subject)
		})
	}
}
