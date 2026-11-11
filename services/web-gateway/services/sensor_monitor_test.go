package services

import (
	"testing"
	"time"

	sensorcontract "lite-nas/shared/contracts/sensormonitor"
	"lite-nas/shared/metrics"
)

// Requirements: sensor-monitor/FR-005.
func TestSensorMonitorSnapshotUsesServiceRPC(t *testing.T) {
	want := metrics.SensorSnapshot{Timestamp: time.Unix(100, 0)}
	client := newSnapshotClientStub(t, want, func(out *sensorcontract.GetSnapshotResponse, snapshot metrics.SensorSnapshot) {
		out.Snapshot = snapshot
	})
	service := NewSensorMonitorService(client)
	mustGetSnapshot(t, service)
	assertMetricsSubject(t, client.subject, sensorcontract.SnapshotRPCSubject)
}

// Requirements: sensor-monitor/FR-005.
func TestSensorMonitorHistoryUsesServiceRPC(t *testing.T) {
	want := []metrics.SensorSnapshot{{Timestamp: time.Unix(100, 0)}}
	client := newHistoryClientStub(t, want, func(out *sensorcontract.GetHistoryResponse, items []metrics.SensorSnapshot) {
		out.Items = items
	})
	service := NewSensorMonitorService(client)
	mustGetHistory(t, service)
	assertMetricsSubject(t, client.subject, sensorcontract.HistoryRPCSubject)
}
