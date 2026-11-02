import { useRbac } from "@hooks/useRbac";
import { SessionsLandingPage } from "@pages/SessionsLandingPage";
import { screen } from "@testing-library/react";
import { renderWithLanguage } from "@tests/unit/test-utils/language";
import { MemoryRouter } from "react-router-dom";

vi.mock("@hooks/useRbac");

/** Mounts the sessions overview with the selected role capabilities. */
const renderPage = (security: boolean, admin: boolean): void => {
  vi.mocked(useRbac).mockReturnValue({
    requireSecurity: () => security,
    requireAdmin: () => admin,
  } as ReturnType<typeof useRbac>);
  renderWithLanguage(
    <MemoryRouter>
      <SessionsLandingPage />
    </MemoryRouter>,
  );
};

describe("SessionsLandingPage", () => {
  test("shows loginctl sessions to security officers", () => {
    renderPage(true, false);
    expect(screen.getByRole("link", { name: /loginctl/i })).toBeInTheDocument();
  });

  test("hides LiteNAS Auth sessions from security officers", () => {
    renderPage(true, false);
    expect(screen.queryByRole("link", { name: /LiteNAS Auth/i })).not.toBeInTheDocument();
  });

  test("shows LiteNAS Auth sessions to administrators", () => {
    renderPage(true, true);
    expect(screen.getByRole("link", { name: /LiteNAS Auth/i })).toBeInTheDocument();
  });

  test("hides the overview without security access", () => {
    renderPage(false, false);
    expect(screen.queryByTestId("sessions-landing-page")).not.toBeInTheDocument();
  });
});
