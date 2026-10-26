package main

import (
	"io"
	"sort"

	"lite-nas/shared/mountinfo"
)

type mountRecord = mountinfo.Record

// parseMountInfo delegates complete snapshot parsing to the shared parser
// used by Antivirus for startup reconciliation and identity checks.
func parseMountInfo(reader io.Reader) (map[uint64]mountRecord, error) {
	return mountinfo.Parse(reader)
}

// changedMounts returns deterministic actions relative to an earlier complete
// snapshot. The first snapshot is a baseline and is handled by the caller.
func changedMounts(before, after map[uint64]mountRecord) []mountChange {
	changes := make([]mountChange, 0)
	for id, previous := range before {
		if current, exists := after[id]; !exists {
			changes = append(changes, mountChange{action: "unmounted", mount: previous})
		} else if current != previous {
			changes = append(changes, mountChange{action: "changed", mount: current})
		}
	}
	for id, current := range after {
		if _, exists := before[id]; !exists {
			changes = append(changes, mountChange{action: "mounted", mount: current})
		}
	}
	sort.Slice(changes, func(i, j int) bool {
		if changes[i].mount.ID == changes[j].mount.ID {
			return changes[i].action < changes[j].action
		}
		return changes[i].mount.ID < changes[j].mount.ID
	})
	return changes
}

// mountChange is one state difference queued for lifecycle publication.
type mountChange struct {
	action string
	mount  mountRecord
}
