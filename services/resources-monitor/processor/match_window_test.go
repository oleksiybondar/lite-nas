package processor

import (
	"context"
	"testing"
	"time"

	servicerules "lite-nas/services/resources-monitor/rules"
	"lite-nas/shared/eventmanager"
	sharedlogger "lite-nas/shared/logger"
)

// TestMatchWindowRequiresFiveSamples verifies a short spike is discarded and
// a later qualifying window publishes the cached samples.
func TestMatchWindowRequiresFiveSamples(t *testing.T) {
	rule := buildMemoryThresholdRule()
	rule.MatchWindow = 10
	rule.MinMatches = 5
	client := &recordingClient{}
	processor := New([]servicerules.Rule{rule}, eventmanager.NewManager(0), client, sharedlogger.NewNop())
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	processor.clock = func() time.Time { return now }
	for index := 0; index < 4; index++ {
		processor.ProcessSnapshot(context.Background(), rule.Event, map[string]any{"snapshot": map[string]any{"mem": map[string]any{"usedPct": 95.0}}})
		now = now.Add(time.Second)
	}
	now = now.Add(11 * time.Second)
	processor.ProcessCooldowns(context.Background())
	for index := 0; index < 5; index++ {
		processor.ProcessSnapshot(context.Background(), rule.Event, map[string]any{"snapshot": map[string]any{"mem": map[string]any{"usedPct": 95.0}}})
		now = now.Add(time.Second)
	}
	if len(client.published) != 6 {
		t.Fatalf("published %d messages, want one alert and five occurrences", len(client.published))
	}
}
