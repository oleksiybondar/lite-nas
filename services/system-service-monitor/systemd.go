package main

import (
	"context"

	servicecontract "lite-nas/shared/contracts/servicemonitor"
	"lite-nas/shared/systemd"
)

type unitState = systemd.UnitState

func validUnitName(unit string) bool { return systemd.ValidUnitName(unit) }
func probeUnit(ctx context.Context, unit string) (unitState, error) {
	return systemd.ProbeUnit(ctx, unit)
}

func parseUnitState(unit, output string) (unitState, error) {
	return systemd.ParseUnitState(unit, output)
}

// classifyUnit maps observed properties to a factual condition. A return to
// active from any unavailable state is a restoration transition.
func classifyUnit(previous *unitState, current unitState) servicecontract.Condition {
	switch {
	case current.ActiveState == "active" && previous != nil && previous.ActiveState != "active":
		return servicecontract.Restored
	case current.ActiveState == "active":
		return servicecontract.Active
	case current.SubState == "auto-restart" || current.ActiveState == "activating":
		return servicecontract.Restarting
	case current.Result == "start-limit-hit":
		return servicecontract.PersistentFailure
	case current.ActiveState == "failed":
		return servicecontract.Failed
	default:
		return servicecontract.Inactive
	}
}
