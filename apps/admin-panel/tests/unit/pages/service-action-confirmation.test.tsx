import { ServiceMetricCard } from "@components/monitoring/ServiceMetricCard";
import { fireEvent, screen, waitFor } from "@testing-library/react";
import { renderWithLanguage } from "@tests/unit/test-utils/language";

const service = { name: "example.service", manager: "systemd", unit_type: "service" };

/** Renders one service with observable action handlers. */
function renderCard(activeState: string) {
  const serviceActions = {
    restart: vi.fn(async () => undefined),
    toggleState: vi.fn(async () => undefined),
    toggleStartupBehaviour: vi.fn(async () => undefined),
  };
  renderWithLanguage(
    <ServiceMetricCard
      service={{ ...service, active_state: activeState }}
      serviceActions={serviceActions}
    />,
  );
  return serviceActions;
}

test("service start waits for button confirmation", async () => {
  const actions = renderCard("inactive");
  fireEvent.click(screen.getByTestId("service-metric-toggle-state-example.service"));
  fireEvent.click(screen.getByTestId("confirm-action-submit"));
  await waitFor(() => expect(actions.toggleState).toHaveBeenCalledWith("example.service", "start"));
});

test("service stop waits for typed yes", async () => {
  const actions = renderCard("active");
  fireEvent.click(screen.getByTestId("service-metric-toggle-state-example.service"));
  fireEvent.change(screen.getByRole("textbox", { name: "Type yes to confirm" }), {
    target: { value: "yes" },
  });
  fireEvent.click(screen.getByTestId("confirm-action-submit"));
  await waitFor(() => expect(actions.toggleState).toHaveBeenCalledWith("example.service", "stop"));
});

test("service restart waits for button confirmation", async () => {
  const actions = renderCard("active");
  fireEvent.click(screen.getByTestId("service-metric-restart-example.service"));
  fireEvent.click(screen.getByTestId("confirm-action-submit"));
  await waitFor(() => expect(actions.restart).toHaveBeenCalledWith("example.service"));
});

test("canceling service stop sends no command", () => {
  const actions = renderCard("active");
  fireEvent.click(screen.getByTestId("service-metric-toggle-state-example.service"));
  fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
  expect(actions.toggleState).not.toHaveBeenCalled();
});
