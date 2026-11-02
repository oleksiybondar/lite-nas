import { useRbac } from "@hooks/useRbac";
import { SecurityLandingPage } from "@pages/SecurityLandingPage";
import { screen } from "@testing-library/react";
import { renderWithLanguage } from "@tests/unit/test-utils/language";
import { MemoryRouter } from "react-router-dom";

vi.mock("@hooks/useRbac");

/** Renders the security overview under the selected security permission. */
const renderPage = (allowed: boolean): void => {
  vi.mocked(useRbac).mockReturnValue({ requireSecurity: () => allowed } as ReturnType<
    typeof useRbac
  >);
  renderWithLanguage(
    <MemoryRouter>
      <SecurityLandingPage />
    </MemoryRouter>,
  );
};

describe("SecurityLandingPage", () => {
  test("links security users to the sessions overview", () => {
    renderPage(true);
    expect(screen.getByRole("link", { name: /sessions/i })).toHaveAttribute(
      "href",
      "/security/sessions",
    );
  });

  test("hides the overview without security access", () => {
    renderPage(false);
    expect(screen.queryByTestId("category-landing-page")).not.toBeInTheDocument();
  });
});
