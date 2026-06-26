import { parseServiceMetricSnapshotResponse } from "@schemas/monitoring/service-metric";

const serviceMetricSnapshotBody = {
  data: {
    services: [
      {
        active_state: "active",
        description: "LiteNAS service metrics service",
        enabled_state: "enabled",
        load_state: "loaded",
        manager: "systemd",
        name: "lite-nas-service-metrics.service",
        pids: [2212640],
        resources: {
          cpu_usage_usec: 697211706,
          memory_bytes: 16809984,
        },
        sub_state: "running",
        unit_type: "service",
      },
    ],
    timestamp: "2026-06-26T04:37:21.316499269+02:00",
  },
  success: true,
  timestamp: "2026-06-26T04:37:22.403092352+02:00",
};

describe("service metrics snapshot schemas", () => {
  test("parses snapshot envelopes and normalizes missing services to an empty list", () => {
    expect(
      parseServiceMetricSnapshotResponse({
        data: {
          services: null,
          timestamp: "2026-06-26T04:37:21.316499269+02:00",
        },
        success: true,
        timestamp: "2026-06-26T04:37:22.403092352+02:00",
      }),
    ).toEqual({
      services: [],
      timestamp: "2026-06-26T04:37:21.316499269+02:00",
    });

    expect(parseServiceMetricSnapshotResponse(serviceMetricSnapshotBody)).toEqual({
      services: [
        {
          active_state: "active",
          description: "LiteNAS service metrics service",
          enabled_state: "enabled",
          load_state: "loaded",
          manager: "systemd",
          name: "lite-nas-service-metrics.service",
          pids: [2212640],
          resources: {
            cpu_usage_usec: 697211706,
            memory_bytes: 16809984,
          },
          sub_state: "running",
          unit_type: "service",
        },
      ],
      timestamp: "2026-06-26T04:37:21.316499269+02:00",
    });
  });
});
