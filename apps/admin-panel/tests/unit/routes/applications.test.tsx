import { ApplicationsLandingPage } from "@pages/ApplicationsLandingPage";
import { FileExplorerEmbeddedPage } from "@pages/FileExplorerEmbeddedPage";
import { applicationsRoutes } from "@routes/applications/routes";
import { appNavigationItems } from "@routes/navigation";
import { render, screen } from "@testing-library/react";
import { TestMemoryRouter } from "@tests/unit/test-utils/router";

describe("Applications navigation", () => {
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
