import { useApi } from "@hooks/useApi";
import { useRbac } from "@hooks/useRbac";
import { AuthSessionsPage } from "@pages/AuthSessionsPage";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { fireEvent, screen } from "@testing-library/react";
import { renderWithLanguage } from "@tests/unit/test-utils/language";

vi.mock("@hooks/useApi");
vi.mock("@hooks/useRbac");

const get = vi.fn(() => ({
  execute: async () =>
    new Response(
      JSON.stringify({
        success: true,
        timestamp: "2026-11-02T00:00:00Z",
        data: [
          {
            session_id: "session_1",
            subject: "1000",
            login: "alice",
            expires_at: "2026-11-02T12:00:00Z",
          },
        ],
      }),
      { status: 200 },
    ),
}));
const post = vi.fn(() => ({
  execute: async () =>
    new Response(
      JSON.stringify({ success: true, timestamp: "2026-11-02T00:00:00Z", revoked: true }),
      { status: 200 },
    ),
}));

/** Mounts the LiteNAS Auth page with isolated API and query state. */
const renderPage = (allowed: boolean): void => {
  vi.mocked(useRbac).mockReturnValue({ requireAdmin: () => allowed } as ReturnType<typeof useRbac>);
  vi.mocked(useApi).mockReturnValue({ get, post } as unknown as ReturnType<typeof useApi>);
  renderWithLanguage(
    <QueryClientProvider
      client={new QueryClient({ defaultOptions: { queries: { retry: false } } })}
    >
      <AuthSessionsPage />
    </QueryClientProvider>,
  );
};

describe("AuthSessionsPage", () => {
  test("does not load sessions without administrator access", () => {
    get.mockClear();
    renderPage(false);
    expect(get).not.toHaveBeenCalled();
  });

  test("shows active LiteNAS Auth sessions", async () => {
    renderPage(true);
    expect(await screen.findByText("alice")).toBeInTheDocument();
  });

  test("revokes the selected session after confirmation", async () => {
    post.mockClear();
    renderPage(true);
    fireEvent.click(await screen.findByRole("button", { name: "Revoke" }));
    fireEvent.change(screen.getByRole("textbox", { name: "Type yes to confirm" }), {
      target: { value: "yes" },
    });
    fireEvent.click(screen.getByTestId("confirm-action-submit"));
    expect(post).toHaveBeenCalledWith("/api/auth/sessions/session_1/revoke");
  });
});
