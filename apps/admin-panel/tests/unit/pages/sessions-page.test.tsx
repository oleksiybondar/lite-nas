import { useApi } from "@hooks/useApi";
import { useRbac } from "@hooks/useRbac";
import { SessionsPage } from "@pages/SessionsPage";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, screen } from "@testing-library/react";
import { renderWithLanguage } from "@tests/unit/test-utils/language";

vi.mock("@hooks/useApi");
vi.mock("@hooks/useRbac");

const inventory = {
  success: true,
  timestamp: "2026-10-31T00:00:00Z",
  data: [
    {
      id: "42",
      uid: "1000",
      user: "alice",
      seat: "seat0",
      leader: "1234",
      class: "user",
      tty: "tty2",
      type: "wayland",
      service: "gdm",
      remote: false,
      remote_host: "",
      display: "",
      desktop: "gnome",
    },
  ],
};
const get = vi.fn(() => ({
  execute: async () => new Response(JSON.stringify(inventory), { status: 200 }),
}));
const post = vi.fn(() => ({
  execute: async () =>
    new Response(
      JSON.stringify({ success: true, timestamp: inventory.timestamp, completed: true }),
      { status: 200 },
    ),
}));

/** Mounts the session page with an isolated query cache and mocked gateway. */
const renderPage = (allowed: boolean): void => {
  vi.mocked(useRbac).mockReturnValue({ requireSecurity: () => allowed } as ReturnType<
    typeof useRbac
  >);
  vi.mocked(useApi).mockReturnValue({ get, post } as unknown as ReturnType<typeof useApi>);
  renderWithLanguage(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <SessionsPage />
    </QueryClientProvider>,
  );
};

describe("SessionsPage", () => {
  test("does not request sessions without security access", () => {
    get.mockClear();
    renderPage(false);
    expect(get).not.toHaveBeenCalled();
  });

  test("renders validated sessions from the sessions endpoint", async () => {
    renderPage(true);
    expect(await screen.findByText("alice")).toBeInTheDocument();
  });

  test("filters sessions by user", async () => {
    renderPage(true);
    await screen.findByText("alice");
    fireEvent.change(screen.getByRole("textbox", { name: "Search sessions" }), {
      target: { value: "bob" },
    });
    expect(screen.getByTestId("sessions-empty")).toBeInTheDocument();
  });

  test("sends a lock action to the sessions endpoint", async () => {
    post.mockClear();
    renderPage(true);
    fireEvent.click(await screen.findByTestId("session-lock-42"));
    fireEvent.click(screen.getByTestId("confirm-action-submit"));
    expect(post).toHaveBeenCalledWith("/api/sessions/42/lock");
  });

  test("requires typed yes before terminating a session", async () => {
    post.mockClear();
    renderPage(true);
    fireEvent.click(await screen.findByTestId("session-terminate-42"));
    fireEvent.change(screen.getByRole("textbox", { name: "Type yes to confirm" }), {
      target: { value: "yes" },
    });
    fireEvent.click(screen.getByTestId("confirm-action-submit"));
    expect(post).toHaveBeenCalledWith("/api/sessions/42/terminate");
  });
});
