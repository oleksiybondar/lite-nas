// @vitest-environment jsdom
import { DashboardPage } from "@pages/DashboardPage";
import { cleanup, fireEvent, screen } from "@testing-library/react";
import { renderWithLanguage as render } from "@tests/unit/test-utils/language";
import { MemoryRouter, Route, Routes } from "react-router-dom";
import { afterEach, expect, test } from "vitest";

afterEach(() => {
  cleanup();
});

test("admin dashboard links change routes without reloading the document", async () => {
  render(
    <MemoryRouter initialEntries={["/"]}>
      <Routes>
        <Route path="/" element={<DashboardPage />} />
        <Route path="/applications" element={<span>Applications route</span>} />
      </Routes>
    </MemoryRouter>,
  );
  fireEvent.click(screen.getByRole("link", { name: /Applications/ }));
  expect(await screen.findByText("Applications route")).toBeInTheDocument();
});
