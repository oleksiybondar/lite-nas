import { usePollingQuery } from "@domain/monitoring/hooks/usePollingQuery";
import type { ServiceMetricSnapshotDTO } from "@dto/monitoring/service-metric";
import { useMomentsAvailability } from "@hooks/useMomentsAvailability";
import type { UseQueryResult } from "@tanstack/react-query";
import { renderHook } from "@testing-library/react";

vi.mock("@domain/monitoring/hooks/usePollingQuery", () => ({ usePollingQuery: vi.fn() }));

describe("useMomentsAvailability", () => {
  // Requirements: moments/FR-005
  test.each([
    { name: "lite-nas-moments.service", sub_state: "running", expected: true },
    { name: "lite-nas-moments.service", sub_state: "dead", expected: false },
    { name: "unrelated.service", sub_state: "running", expected: false },
  ])("reports $expected for $name in $sub_state state", ({ name, sub_state, expected }) => {
    const snapshot: ServiceMetricSnapshotDTO = {
      services: [{ manager: "system", name, sub_state, unit_type: "service" }],
      timestamp: "2026-10-23T00:00:00Z",
    };
    vi.mocked(usePollingQuery).mockReturnValue({ data: snapshot } as UseQueryResult<
      ServiceMetricSnapshotDTO,
      Error
    >);

    expect(renderHook(() => useMomentsAvailability()).result.current).toBe(expected);
  });

  // Requirements: moments/FR-005
  test("hides Moments when service status is unavailable", () => {
    vi.mocked(usePollingQuery).mockReturnValue({ data: undefined } as UseQueryResult<
      ServiceMetricSnapshotDTO,
      Error
    >);

    expect(renderHook(() => useMomentsAvailability()).result.current).toBe(false);
  });

  // Requirements: moments/FR-005
  test("hides Moments when the latest status request fails", () => {
    vi.mocked(usePollingQuery).mockReturnValue({
      data: {
        services: [
          {
            manager: "system",
            name: "lite-nas-moments.service",
            sub_state: "running",
            unit_type: "service",
          },
        ],
        timestamp: "2026-10-23T00:00:00Z",
      },
      isError: true,
    } as UseQueryResult<ServiceMetricSnapshotDTO, Error>);

    expect(renderHook(() => useMomentsAvailability()).result.current).toBe(false);
  });
});
