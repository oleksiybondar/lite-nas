import { useApi } from "@hooks/useApi";
import { useRbac } from "@hooks/useRbac";
import { AutomatedAuditPage } from "@pages/AutomatedAuditPage";
import { AutomatedAuditProvider } from "@providers/AutomatedAuditProvider";
import { fireEvent, screen } from "@testing-library/react";
import { renderWithLanguage } from "@tests/unit/test-utils/language";

vi.mock("@hooks/useApi");
vi.mock("@hooks/useRbac");

const get = vi.fn(() => ({
  execute: async () =>
    new Response(
      JSON.stringify({
        data: {
          rules: [
            {
              id: "login_failures",
              name: "Login failures",
              explanation: "Detects repeated login failures.",
              possible_reactions: null,
              enabled: true,
              source: "journal",
              category: "authentication",
              severity: "high",
              instance: "auth",
              threshold: 5,
              window: "1m",
              email_on_critical: false,
              lockdown: false,
            },
          ],
        },
      }),
      { status: 200 },
    ),
}));

/** Mounts the audit catalog with an authenticated role and gateway response. */
const renderPage = (allowed: boolean): void => {
  vi.mocked(useRbac).mockReturnValue({ requireSecurity: () => allowed } as ReturnType<
    typeof useRbac
  >);
  vi.mocked(useApi).mockReturnValue({ get } as unknown as ReturnType<typeof useApi>);
  renderWithLanguage(
    <AutomatedAuditProvider>
      <AutomatedAuditPage />
    </AutomatedAuditProvider>,
  );
};

describe("AutomatedAuditPage", () => {
  test("does not request rules without security access", () => {
    get.mockClear();
    renderPage(false);
    expect(get).not.toHaveBeenCalled();
  });

  test("shows a rule with a nullable reactions list", async () => {
    renderPage(true);
    expect(await screen.findByText("Detects repeated login failures.")).toBeInTheDocument();
  });

  test("has no rule action buttons", async () => {
    renderPage(true);
    await screen.findByText("Login failures");
    expect(
      screen.queryByRole("button", { name: /disable|mitigate|rollback/i }),
    ).not.toBeInTheDocument();
  });

  test("filters by explanation and highlights the matching text", async () => {
    renderPage(true);
    await screen.findByText("Login failures");
    fireEvent.change(screen.getByRole("textbox", { name: "Search audit rules" }), {
      target: { value: "repeated" },
    });
    expect(screen.getByText("repeated", { selector: "mark" })).toBeInTheDocument();
  });

  test("shows a distinct empty result for an unmatched search", async () => {
    renderPage(true);
    await screen.findByText("Login failures");
    fireEvent.change(screen.getByRole("textbox", { name: "Search audit rules" }), {
      target: { value: "unrelated" },
    });
    expect(screen.getByText("No audit rules match your search.")).toBeInTheDocument();
  });
});
