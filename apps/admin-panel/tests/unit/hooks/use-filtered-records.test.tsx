import { useFilteredRecords } from "@hooks/useFilteredRecords";
import { act, renderHook } from "@testing-library/react";

type ServiceRecord = {
  activeState: string;
  description: string;
  enabledState: string;
  name: string;
};

const serviceRecords: ServiceRecord[] = [
  {
    activeState: "active",
    description: "LiteNAS service metrics service",
    enabledState: "enabled",
    name: "lite-nas-service-metrics.service",
  },
  {
    activeState: "inactive",
    description: "OpenBSD Secure Shell server",
    enabledState: "disabled",
    name: "ssh.service",
  },
  {
    activeState: "active",
    description: "Network Manager",
    enabledState: "enabled",
    name: "NetworkManager.service",
  },
];

describe("useFilteredRecords", () => {
  test("filters records by free-text search and named filters", () => {
    const { result } = renderHook(() =>
      useFilteredRecords<ServiceRecord, "activeState" | "enabledState">({
        getSearchText: (record) => `${record.name} ${record.description}`,
        initialFilters: {
          activeState: "",
          enabledState: "",
        },
        matchesFilter: (record, key, value) => {
          if (key === "activeState") {
            return record.activeState === value;
          }

          return record.enabledState === value;
        },
        records: serviceRecords,
      }),
    );

    expect(result.current.records.map((record) => record.name)).toEqual([
      "lite-nas-service-metrics.service",
      "ssh.service",
      "NetworkManager.service",
    ]);
    expect(result.current.totalCount).toBe(3);
    expect(result.current.filteredCount).toBe(3);

    act(() => {
      result.current.setSearch("network");
    });

    expect(result.current.records.map((record) => record.name)).toEqual(["NetworkManager.service"]);

    act(() => {
      result.current.setSearch("");
      result.current.setFilterValue("activeState", "active");
      result.current.setFilterValue("enabledState", "enabled");
    });

    expect(result.current.records.map((record) => record.name)).toEqual([
      "lite-nas-service-metrics.service",
      "NetworkManager.service",
    ]);

    act(() => {
      result.current.resetFilters();
    });

    expect(result.current.search).toBe("");
    expect(result.current.filters).toEqual({ activeState: "", enabledState: "" });
    expect(result.current.records).toHaveLength(3);
  });
});
