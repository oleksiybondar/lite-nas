import { useRbac } from "@hooks/useRbac";
import { AutomationPage } from "@pages/AutomationPage";
import { AutomationProvider } from "@providers/AutomationProvider";
import { useAutomation } from "@providers/useAutomation";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import { renderWithLanguage } from "@tests/unit/test-utils/language";

vi.mock("@hooks/useRbac");
vi.mock("@providers/useAutomation");

/** Renders one explained rule and one active incident. */
const renderAutomation = (resetCooldown = vi.fn()): void => {
  vi.mocked(useRbac).mockReturnValue({ requireSecurity: () => true } as ReturnType<typeof useRbac>);
  const rule = {
    id: "scan",
    name: "Scan response",
    description: "Detect repeated scans",
    explanation: "Protects against repeated scans from one address.",
    enabled: true,
    configured_enabled: true,
    action_kind: "builtin" as const,
    action_name: "block_source_ip",
    state: "mitigating" as const,
    active_count: 1,
    last_outcome: "",
  };
  vi.mocked(useAutomation).mockReturnValue({
    rules: [rule],
    active: [
      {
        incident_id: "incident1",
        rule,
        target: "192.0.2.1",
        mitigation: "block_source_ip",
        started_at: "2026-11-02T00:00:00Z",
        state: "mitigating",
        last_outcome: "applied",
      },
    ],
    loading: false,
    error: false,
    outcome: "",
    refresh: vi.fn(),
    setEnabled: vi.fn(),
    rollback: vi.fn(),
    forceMitigate: vi.fn(),
    resetCooldown,
  });
  renderWithLanguage(
    <AutomationProvider>
      <AutomationPage />
    </AutomationProvider>,
  );
};

// biome-ignore lint/complexity/noExcessiveLinesPerFunction: These focused cases share one page fixture.
describe("AutomationPage", () => {
  test("shows the configured rule explanation", () => {
    renderAutomation();
    expect(
      screen.getAllByText("Protects against repeated scans from one address."),
    ).not.toHaveLength(0);
  });

  test("filters rules by explanation and highlights the match", () => {
    renderAutomation();
    fireEvent.change(screen.getByRole("textbox", { name: "Search rules and active automations" }), {
      target: { value: "Protects" },
    });
    expect(screen.getAllByText("Protects", { selector: "mark" })).not.toHaveLength(0);
  });

  test("filters active mitigations by target", () => {
    renderAutomation();
    fireEvent.change(screen.getByRole("textbox", { name: "Search rules and active automations" }), {
      target: { value: "192.0.2.1" },
    });
    expect(screen.getByText("192.0.2.1", { selector: "mark" })).toBeInTheDocument();
  });

  test("finds a rule by its action name", () => {
    renderAutomation();
    fireEvent.change(screen.getByRole("textbox", { name: "Search rules and active automations" }), {
      target: { value: "block_source_ip" },
    });
    expect(screen.getByText("Scan response")).toBeInTheDocument();
  });

  test("shows the active incident rollback command", () => {
    renderAutomation();
    expect(screen.getByText("security-automation-cli rollback incident1")).toBeInTheDocument();
  });

  test("requires typed confirmation before rolling back an incident", () => {
    renderAutomation();
    fireEvent.click(screen.getByRole("button", { name: "Roll back mitigation" }));
    expect(screen.getByRole("button", { name: /continue/i })).toBeDisabled();
  });

  test("requires a correlation key before resetting detection state", () => {
    renderAutomation();
    fireEvent.click(screen.getByRole("button", { name: "Reset detection state" }));
    fireEvent.change(screen.getByRole("textbox", { name: "Type yes to confirm" }), {
      target: { value: "yes" },
    });
    expect(screen.getByRole("button", { name: /continue/i })).toBeDisabled();
  });

  test("sends the selected rule and correlation key for reset", async () => {
    const resetCooldown = vi.fn();
    renderAutomation(resetCooldown);
    fireEvent.click(screen.getByRole("button", { name: "Reset detection state" }));
    fireEvent.change(screen.getByRole("textbox", { name: "Correlation key" }), {
      target: { value: "192.0.2.1" },
    });
    fireEvent.change(screen.getByRole("textbox", { name: "Type yes to confirm" }), {
      target: { value: "yes" },
    });
    fireEvent.click(screen.getByRole("button", { name: /continue/i }));
    await waitFor(() => expect(resetCooldown).toHaveBeenCalledWith("scan", "192.0.2.1"));
  });
});
