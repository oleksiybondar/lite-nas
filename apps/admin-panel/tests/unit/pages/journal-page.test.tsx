import { useApi } from "@hooks/useApi";
import { useRbac } from "@hooks/useRbac";
import { JournalPage } from "@pages/JournalPage";
import { fireEvent, screen } from "@testing-library/react";
import { renderWithLanguage } from "@tests/unit/test-utils/language";

vi.mock("@hooks/useApi");
vi.mock("@hooks/useRbac");

const chunk = {
  success: true,
  data: {
    entries: [
      {
        cursor: "s=one",
        timestamp: "1793558018131899",
        message: "Disk ready",
        unit: "disk.service",
      },
    ],
    next_cursor: "s=one",
    has_more: true,
  },
};
const get = vi.fn((path: string) => ({
  execute: async () =>
    new Response(
      JSON.stringify(
        path.startsWith("/api/journal/units")
          ? { success: true, data: { units: ["disk.service", "ssh.service"] } }
          : chunk,
      ),
      { status: 200 },
    ),
}));
const post = vi.fn(() => ({
  execute: async () =>
    new Response(JSON.stringify({ success: true, completed: true }), { status: 200 }),
}));

/** Mounts the page with the requested security role and mocked gateway. */
const renderPage = (allowed: boolean): void => {
  vi.mocked(useRbac).mockReturnValue({ requireSecurity: () => allowed } as ReturnType<
    typeof useRbac
  >);
  vi.mocked(useApi).mockReturnValue({ get, post } as unknown as ReturnType<typeof useApi>);
  renderWithLanguage(<JournalPage />);
};

describe("JournalPage", () => {
  test("hides journal controls without security access", () => {
    renderPage(false);
    expect(screen.queryByTestId("journal-browse")).not.toBeInTheDocument();
  });

  test("loads structured entries into the modal", async () => {
    renderPage(true);
    fireEvent.click(screen.getByTestId("journal-browse"));
    expect(await screen.findByText("Disk ready")).toBeInTheDocument();
  });

  test("marks literal message matches in loaded entries", async () => {
    renderPage(true);
    fireEvent.click(screen.getByTestId("journal-browse"));
    await screen.findByText("Disk ready");
    fireEvent.change(screen.getByRole("textbox", { name: "Search message" }), {
      target: { value: "disk" },
    });
    expect(screen.getByText("Disk", { selector: "mark" })).toBeInTheDocument();
  });

  test("queries the selected matching unit", async () => {
    renderPage(true);
    fireEvent.click(screen.getByTestId("journal-browse"));
    await screen.findByText("Disk ready");
    fireEvent.mouseDown(screen.getByRole("combobox", { name: "Systemd unit" }));
    fireEvent.click(screen.getByRole("option", { name: "disk.service" }));
    fireEvent.click(screen.getByTestId("journal-apply-filters"));
    await screen.findByText("Disk ready");
    expect(get).toHaveBeenCalledWith(expect.stringContaining("unit=disk.service"));
  });
});
