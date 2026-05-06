import {
  appNavigationItems,
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
    { pathname: "/preferences", selectedPath: null },
    { pathname: "/missing", selectedPath: null },
  ])("maps $pathname to $selectedPath", ({ pathname, selectedPath }) => {
    expect(resolveSelectedNavigationPath(pathname)).toBe(selectedPath);
  });
});

describe("resolveNavigationItems", () => {
  test("uses preferences navigation for preferences routes", () => {
    expect(resolveNavigationItems("/preferences/profile")).toBe(preferencesNavigationItems);
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
