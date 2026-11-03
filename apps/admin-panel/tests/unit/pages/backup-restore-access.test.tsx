import { useApi } from "@hooks/useApi";
import { useRbac } from "@hooks/useRbac";
import { BackupRestorePage } from "@pages/BackupRestorePage";
import { buildAppNavigationItems } from "@routes/navigation";
import { screen } from "@testing-library/react";
import { renderWithLanguage } from "@tests/unit/test-utils/language";
import { MemoryRouter, Route, Routes } from "react-router-dom";

vi.mock("@hooks/useApi");
vi.mock("@hooks/useRbac");

/** Opens the protected page as a non-admin and returns the API read spy. */
const renderNonAdmin = (): ReturnType<typeof vi.fn> => {
  const get = vi.fn();
  vi.mocked(useApi).mockReturnValue({ get } as unknown as ReturnType<typeof useApi>);
  vi.mocked(useRbac).mockReturnValue({ requireAdmin: () => false } as ReturnType<typeof useRbac>);
  renderWithLanguage(
    <MemoryRouter initialEntries={["/backup-restore"]}>
      <Routes>
        <Route path="/backup-restore" element={<BackupRestorePage />} />
        <Route path="/" element={<div>Dashboard</div>} />
      </Routes>
    </MemoryRouter>,
  );
  return get;
};

describe("backup and restore administrator access", () => {
  test("shows the sidebar entry to administrators", () => {
    const items = buildAppNavigationItems({
      requireAdmin: () => true,
      requireOperator: () => true,
      requireSecurity: () => true,
    });
    expect(items.some((item) => item.path === "/backup-restore")).toBe(true);
  });

  test("hides the sidebar entry from security officers", () => {
    const items = buildAppNavigationItems({
      requireAdmin: () => false,
      requireOperator: () => false,
      requireSecurity: () => true,
    });
    expect(items.some((item) => item.path === "/backup-restore")).toBe(false);
  });

  test("redirects a non-admin who opens the URL directly", async () => {
    renderNonAdmin();
    expect(await screen.findByText("Dashboard")).toBeInTheDocument();
  });

  test("does not fetch backup data for a non-admin", () => {
    const get = renderNonAdmin();
    expect(get).not.toHaveBeenCalled();
  });
});
