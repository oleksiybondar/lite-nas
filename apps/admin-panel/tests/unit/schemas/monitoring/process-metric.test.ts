import { parseProcessMetricSnapshotResponse } from "@schemas/monitoring/process-metric";
import { expectMonitoringSnapshotParser } from "@tests/unit/schemas/monitoring/helpers";

const processMetricSnapshotBody = {
  data: {
    processes: [
      {
        cmdline: "/usr/libexec/lite-nas/process-metrics",
        cpu: {
          system_ticks: 121,
          total_ticks: 194,
          user_ticks: 73,
        },
        gid: 958,
        memory: {
          rss_bytes: 12283904,
          vms_bytes: 1266016256,
        },
        name: "process-metrics",
        open_fds: 9,
        pid: 414669,
        ppid: 1,
        start_time: "2026-06-24T18:57:24.46Z",
        state: "S",
        threads: 10,
        uid: 971,
        username: "lite-nas-process-metrics",
      },
    ],
    timestamp: "2026-06-24T21:03:20.452436529+02:00",
  },
  success: true,
  timestamp: "2026-06-24T21:03:24.62464624+02:00",
};

describe("process metrics snapshot schemas", () => {
  test("parses snapshot envelopes and normalizes missing processes to an empty list", () => {
    expectMonitoringSnapshotParser({
      emptyEnvelope: {
        data: {
          processes: null,
          timestamp: "2026-06-24T21:03:20.452436529+02:00",
        },
        success: true,
        timestamp: "2026-06-24T21:03:24.62464624+02:00",
      },
      expectedEmptyResult: {
        processes: [],
        timestamp: "2026-06-24T21:03:20.452436529+02:00",
      },
      expectedParsedResult: {
        processes: processMetricSnapshotBody.data.processes,
        timestamp: "2026-06-24T21:03:20.452436529+02:00",
      },
      parser: parseProcessMetricSnapshotResponse,
      populatedEnvelope: processMetricSnapshotBody,
    });
  });
});
