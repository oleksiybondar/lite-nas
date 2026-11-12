import { collectAdminSettings } from "@helpers/profile-settings";
import { useApi } from "@hooks/useApi";
import { useSensorMarkers } from "@hooks/useSensorMarkers";
import { act, renderHook } from "@testing-library/react";

vi.mock("@hooks/useApi");

test("saving one chart migrates valid legacy stops and ignores a malformed old entry", async () => {
  const profile = {
    ...collectAdminSettings(),
    sensorMarkers: {
      "voltages:core": [{ value: 5, color: "#00ff00" }],
      "voltages:broken": [{ value: 5, color: "green" }],
    },
  };
  const get = vi.fn(() => ({
    execute: async () => new Response(JSON.stringify(profile), { status: 200 }),
  }));
  const put = vi.fn((_url: string, _payload: unknown) => ({
    execute: async () => new Response(null, { status: 200 }),
  }));
  vi.mocked(useApi).mockReturnValue({ get, put } as unknown as ReturnType<typeof useApi>);
  const { result } = renderHook(() => useSensorMarkers());

  await act(async () => {
    await result.current.saveChart("temperatures:cpu", {
      mode: "regular",
      stops: [{ value: 70, color: "#ff0000", label: "Hot", showBar: true, showLabel: true }],
    });
  });

  const payload = put.mock.calls[0]?.[1] as { document: { sensorCharts: unknown } };
  expect(payload.document.sensorCharts).toEqual({
    "voltages:core": {
      mode: "gradient",
      stops: [{ value: 5, color: "#00ff00", label: "", showBar: false, showLabel: false }],
    },
    "temperatures:cpu": {
      mode: "regular",
      stops: [{ value: 70, color: "#ff0000", label: "Hot", showBar: true, showLabel: true }],
    },
  });
});
