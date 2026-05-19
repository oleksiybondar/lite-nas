package snapshot

import (
	"testing"
	"time"

	iostatparser "lite-nas/shared/parsers/zfs/iostat"
	listparser "lite-nas/shared/parsers/zfs/list"
	statusparser "lite-nas/shared/parsers/zfs/status"
)

func TestCompose(t *testing.T) {
	doc := statusparser.StatusDocument{
		Pools: []statusparser.PoolBlock{
			{
				PoolName:      "LiteNAS",
				ErrorsSummary: "No known data errors",
				Metadata: statusparser.PoolMetadata{
					State: "ONLINE",
					Scan:  "scan text",
				},
				Config: statusparser.ConfigTree{
					Roots: []statusparser.ConfigNode{
						{
							Name: "LiteNAS",
							Columns: map[string]string{
								"READ":  "0",
								"WRITE": "0",
								"CKSUM": "0",
							},
						},
					},
				},
			},
		},
	}

	usageByPool := map[string]listparser.PoolUsage{
		"LiteNAS": {CapacityPct: 1},
	}
	ioByPool := map[string]iostatparser.PoolIOStat{
		"LiteNAS": {Bandwidth: iostatparser.IOStatValues{Read: 44544}},
	}

	snapshot := Compose(time.Unix(100, 0), doc, usageByPool, ioByPool)
	if len(snapshot.Pools) != 1 {
		t.Fatalf("expected one pool, got %d", len(snapshot.Pools))
	}
	if snapshot.Pools[0].Usage == nil || snapshot.Pools[0].IOStat == nil {
		t.Fatal("expected enriched usage and iostat")
	}
}
