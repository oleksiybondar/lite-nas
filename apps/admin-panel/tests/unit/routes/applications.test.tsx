import { useMomentsAvailability } from "@hooks/useMomentsAvailability";
import { ApplicationsLandingPage } from "@pages/ApplicationsLandingPage";
import { FileExplorerEmbeddedPage } from "@pages/FileExplorerEmbeddedPage";
import { MomentsEmbeddedPage } from "@pages/MomentsEmbeddedPage";
import { applicationsRoutes } from "@routes/applications/routes";
import { appNavigationItems, buildAppNavigationItems } from "@routes/navigation";
import { render, screen } from "@testing-library/react";
import { TestMemoryRouter } from "@tests/unit/test-utils/router";

vi.mock("@hooks/useMomentsAvailability", () => ({ useMomentsAvailability: vi.fn() }));

describe("File Explorer application", () => {
  test("lists File Explorer under the Applications category", () => {
    const category = appNavigationItems.find((item) => item.path === "/applications");

    expect(category?.children?.[0]?.path).toBe("/applications/file-explorer");
  });

  test("keeps the embedded route outside the Nginx app mount", () => {
    expect(applicationsRoutes[1]?.path).toBe("/applications/file-explorer");
  });

  test("loads the separately served File Explorer in an iframe", () => {
    render(<FileExplorerEmbeddedPage />);

    expect(screen.getByTitle("File Explorer")).toHaveAttribute("src", "/apps/file-explorer/");
  });

  test("admin-panel-app/FR-001 opens File Explorer from the whole Applications card", () => {
    render(
      <TestMemoryRouter>
        <ApplicationsLandingPage />
      </TestMemoryRouter>,
    );

    expect(screen.getByRole("link", { name: /File Explorer/ })).toHaveAttribute(
      "href",
      "/applications/file-explorer",
    );
  });
});

describe("Moments application", () => {
  // Requirements: moments/FR-005
  test("lists Moments under the Applications category while its service is running", () => {
    const category = buildAppNavigationItems(
      { requireOperator: () => false, requireSecurity: () => false },
      true,
    ).find((item) => item.path === "/applications");

    expect(category?.children?.some((item) => item.path === "/applications/moments")).toBe(true);
  });

  // Requirements: moments/FR-005
  test("hides Moments from Applications navigation while its service is stopped", () => {
    const category = appNavigationItems.find((item) => item.path === "/applications");

    expect(category?.children?.some((item) => item.path === "/applications/moments")).toBe(false);
  });

  // Requirements: moments/FR-005
  test("keeps the Moments admin route outside the Nginx app mount", () => {
    expect(applicationsRoutes[2]?.path).toBe("/applications/moments");
  });

  // Requirements: moments/FR-005
  test("loads the separately served Moments app in an iframe", () => {
    render(<MomentsEmbeddedPage />);

    expect(screen.getByTitle("Moments")).toHaveAttribute("src", "/apps/moments/");
  });

  // Requirements: moments/FR-005
  test("opens Moments from its Applications card", () => {
    vi.mocked(useMomentsAvailability).mockReturnValue(true);
    render(
      <TestMemoryRouter>
        <ApplicationsLandingPage />
      </TestMemoryRouter>,
    );

    expect(screen.getByRole("link", { name: /Moments/ })).toHaveAttribute(
      "href",
      "/applications/moments",
    );
  });

  // Requirements: moments/FR-005
  test("hides the Moments card while its service is stopped", () => {
    vi.mocked(useMomentsAvailability).mockReturnValue(false);
    render(
      <TestMemoryRouter>
        <ApplicationsLandingPage />
      </TestMemoryRouter>,
    );

    expect(screen.queryByRole("link", { name: /Moments/ })).not.toBeInTheDocument();
  });
});
