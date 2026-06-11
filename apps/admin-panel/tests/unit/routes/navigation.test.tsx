import {
  appNavigationItems,
  buildAppNavigationItems,
  preferencesNavigationItems,
  resolveNavigationItems,
  resolveSelectedNavigationPath,
} from "@routes/navigation";

describe("resolveSelectedNavigationPath", () => {
  test.each([
    { pathname: "/", selectedPath: "/" },
    { pathname: "/system", selectedPath: "/system" },
    { pathname: "/system/performance", selectedPath: "/system/performance" },
    { pathname: "/system/sensors", selectedPath: "/system/sensors" },
    { pathname: "/system/performance/network", selectedPath: "/system/performance/network" },
    { pathname: "/system/sensors/temperature", selectedPath: "/system/sensors/temperature" },
    {
      pathname: "/system/sensors/temperature/history",
      selectedPath: "/system/sensors/temperature",
    },
    {
      pathname: "/preferences/application/theme",
      selectedPath: "/preferences/application/theme",
      items: preferencesNavigationItems,
    },
    {
      pathname: "/preferences/application/monitoring",
      selectedPath: "/preferences/application/monitoring",
      items: preferencesNavigationItems,
    },
    { pathname: "/preferences", selectedPath: null },
    { pathname: "/missing", selectedPath: null },
  ])("maps $pathname to $selectedPath", ({ items, pathname, selectedPath }) => {
    expect(resolveSelectedNavigationPath(pathname, items)).toBe(selectedPath);
  });
});

describe("resolveNavigationItems", () => {
  test("uses preferences navigation for preferences routes", () => {
    expect(resolveNavigationItems("/preferences/profile")).toBe(preferencesNavigationItems);
  });

  test("keeps application settings children in preferences navigation", () => {
    const routePaths = JSON.stringify(resolveNavigationItems("/preferences/application/theme"));

    expect(routePaths).toContain("/preferences/application/theme");
    expect(routePaths).toContain("/preferences/application/monitoring");
  });

  test("uses admin navigation for dashboard routes", () => {
    expect(resolveNavigationItems("/system")).toBe(appNavigationItems);
  });
});

describe("appNavigationItems", () => {
  test("contains Raspberry Pi sensor routes", () => {
    const routePaths = JSON.stringify(appNavigationItems);

    expect(routePaths).toContain("/system/sensors/temperature");
    expect(routePaths).toContain("/system/sensors/voltage");
    expect(routePaths).toContain("/system/sensors/clock");
    expect(routePaths).toContain("/system/sensors/throttling");
    expect(routePaths).toContain("/system/sensors/fan");
  });
});

describe("buildAppNavigationItems", () => {
  test("adds only the operator alerts branch for operator visibility", () => {
    const routePaths = JSON.stringify(
      buildAppNavigationItems({
        requireOperator: () => true,
        requireSecurity: () => false,
      }),
    );

    expect(routePaths).toContain("/alerts/system/unacknowledged");
    expect(routePaths).not.toContain("/alerts/security/unacknowledged");
  });

  test("adds both alerts branches when both guards pass", () => {
    const routePaths = JSON.stringify(
      buildAppNavigationItems({
        requireOperator: () => true,
        requireSecurity: () => true,
      }),
    );

    expect(routePaths).toContain("/alerts/system/active");
    expect(routePaths).toContain("/alerts/security/active");
    expect(routePaths).toContain("/alerts/system/all");
    expect(routePaths).toContain("/alerts/security/all");
  });
});
