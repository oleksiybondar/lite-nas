// @vitest-environment jsdom
import { act, cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { afterEach, beforeEach, expect, test, vi } from "vitest";
import { GalleryScanControls } from "./GalleryScanControls";
import {
  cancelGalleryScan,
  getGalleryScan,
  latestGalleryScan,
  startCaptionScan,
  startGalleryScan,
} from "./moments-data";
import type { GalleryScan } from "./schemas";

vi.mock("./moments-data", () => ({
  cancelGalleryScan: vi.fn(),
  getGalleryScan: vi.fn(),
  latestGalleryScan: vi.fn(),
  startCaptionScan: vi.fn(),
  startGalleryScan: vi.fn(),
}));

const scan: GalleryScan = {
  id: "00000000-0000-4000-8000-000000000002",
  galleryId: "00000000-0000-4000-8000-000000000001",
  kind: "MEDIA",
  status: "PENDING",
  totalItems: null,
  processedItems: 0,
  successfulItems: 0,
  failedItems: 0,
  createdAt: "2026-10-26T00:00:00Z",
  startedAt: null,
  finishedAt: null,
  errorSummary: null,
};

/** Waits until the owner action can be submitted after history loads. */
async function readyStartButton(): Promise<HTMLButtonElement> {
  const button = screen.getByRole("button", { name: "Scan media" }) as HTMLButtonElement;
  await waitFor(() => {
    if (button.disabled) throw new Error("Scan history is still loading.");
  });
  return button;
}

beforeEach(() => {
  vi.mocked(latestGalleryScan).mockResolvedValue(null);
  vi.mocked(startGalleryScan).mockResolvedValue(scan);
  vi.mocked(startCaptionScan).mockResolvedValue({ ...scan, kind: "CAPTION" });
  vi.mocked(getGalleryScan).mockResolvedValue(scan);
  vi.mocked(cancelGalleryScan).mockResolvedValue(scan);
});

afterEach(() => {
  cleanup();
  vi.useRealTimers();
  vi.clearAllMocks();
});

test("scan control starts a full scan for the saved gallery", async () => {
  render(<GalleryScanControls galleryId={scan.galleryId} open />);
  fireEvent.click(await readyStartButton());
  await waitFor(() => expect(startGalleryScan).toHaveBeenCalledWith(scan.galleryId));
});

test("caption control requests missing captions by default", async () => {
  render(<GalleryScanControls galleryId={scan.galleryId} open kind="CAPTION" />);
  const button = screen.getByRole("button", { name: "Generate captions" }) as HTMLButtonElement;
  await waitFor(() => expect(button.disabled).toBe(false));
  fireEvent.click(button);
  await waitFor(() => expect(startCaptionScan).toHaveBeenCalledWith(scan.galleryId, false));
});

test("caption control forwards explicit regeneration", async () => {
  render(<GalleryScanControls galleryId={scan.galleryId} open kind="CAPTION" />);
  fireEvent.click(screen.getByRole("switch", { name: "Regenerate existing captions" }));
  const button = screen.getByRole("button", { name: "Generate captions" }) as HTMLButtonElement;
  await waitFor(() => expect(button.disabled).toBe(false));
  fireEvent.click(button);
  await waitFor(() => expect(startCaptionScan).toHaveBeenCalledWith(scan.galleryId, true));
});

test("scan control shows the latest completed job", async () => {
  vi.mocked(latestGalleryScan).mockResolvedValue({ ...scan, status: "COMPLETED" });
  render(<GalleryScanControls galleryId={scan.galleryId} open />);
  expect((await screen.findByRole("status")).textContent).toContain("completed");
});

test("scan control shows measured progress after discovery", async () => {
  vi.mocked(latestGalleryScan).mockResolvedValue({
    ...scan,
    status: "RUNNING",
    totalItems: 10,
    processedItems: 5,
  });
  render(<GalleryScanControls galleryId={scan.galleryId} open />);
  expect((await screen.findByRole("progressbar")).getAttribute("aria-valuenow")).toBe("50");
});

test("scan control restores an active job while the editor is closed", async () => {
  vi.mocked(latestGalleryScan).mockResolvedValue(scan);
  render(<GalleryScanControls galleryId={scan.galleryId} open={false} />);
  expect(await screen.findByRole("button", { name: "Cancel scan" })).toBeTruthy();
});

test("scan control offers cancellation for an active job", async () => {
  vi.mocked(latestGalleryScan).mockResolvedValue(scan);
  render(<GalleryScanControls galleryId={scan.galleryId} open />);
  fireEvent.click(await screen.findByRole("button", { name: "Cancel scan" }));
  await waitFor(() => expect(cancelGalleryScan).toHaveBeenCalledWith(scan.id));
});

test("scan control polls an active job to renew its worker token", async () => {
  vi.useFakeTimers();
  vi.mocked(latestGalleryScan).mockResolvedValue(scan);
  await act(async () => {
    render(<GalleryScanControls galleryId={scan.galleryId} open />);
  });
  await act(async () => {
    vi.advanceTimersByTime(3000);
  });
  expect(getGalleryScan).toHaveBeenCalledWith(scan.id);
});

test("scan control continues polling when the editor closes", async () => {
  vi.useFakeTimers();
  vi.mocked(latestGalleryScan).mockResolvedValue(scan);
  let rerender: ReturnType<typeof render>["rerender"] = () => undefined;
  await act(async () => {
    rerender = render(<GalleryScanControls galleryId={scan.galleryId} open />).rerender;
  });
  rerender(<GalleryScanControls galleryId={scan.galleryId} open={false} />);
  await act(async () => {
    vi.advanceTimersByTime(3000);
  });
  expect(getGalleryScan).toHaveBeenCalledWith(scan.id);
});

test("scan control reloads the latest job when the editor reopens", async () => {
  const view = render(<GalleryScanControls galleryId={scan.galleryId} open />);
  await readyStartButton();
  view.rerender(<GalleryScanControls galleryId={scan.galleryId} open={false} />);
  view.rerender(<GalleryScanControls galleryId={scan.galleryId} open />);
  await waitFor(() => expect(latestGalleryScan).toHaveBeenCalledTimes(2));
});

test("scan control reports a failed start", async () => {
  vi.mocked(startGalleryScan).mockRejectedValue(new Error("unavailable"));
  render(<GalleryScanControls galleryId={scan.galleryId} open />);
  fireEvent.click(await readyStartButton());
  expect(await screen.findByText("Could not start media scan.")).toBeTruthy();
});
