// @vitest-environment jsdom

import { createAppTheme } from "@lite-nas/theme";
import { ThemeProvider } from "@mui/material/styles";
import { act, cleanup, fireEvent, screen, waitFor, within } from "@testing-library/react";
import { afterEach, expect, test, vi } from "vitest";
import { Gallery } from "./Gallery";
import { type GalleryItem, galleryItemsFromPaths } from "./gallery-model";
import { StreamProvider } from "./StreamContext";
import { renderWithLanguage as render } from "./test-utils/language";

const identifier = "A4976DD1-CFD7-4302-9602-FD5B546987BA";

/** Returns the bounded media range shape supplied by File Service. */
function mediaRange(text: string, suffix: boolean): Response {
  const bytes = new TextEncoder().encode(text);
  const start = suffix ? 100 : 0;
  const end = start + bytes.byteLength - 1;
  const total = suffix ? end + 1 : end + 100;
  return new Response(bytes, {
    status: 206,
    headers: {
      "Content-Length": String(bytes.byteLength),
      "Content-Range": `bytes ${start}-${end}/${total}`,
    },
  });
}

/** Reports a stable wall width so the virtual gallery renders its first tile. */
class GalleryResizeObserver {
  private readonly callback: ResizeObserverCallback;

  constructor(callback: ResizeObserverCallback) {
    this.callback = callback;
  }

  observe(): void {
    this.callback([], this as unknown as ResizeObserver);
  }

  disconnect(): void {}
}

/** Renders the same two authorized originals the demo fixture presents. */
function renderLivePhotoGallery(
  items: readonly GalleryItem[] = galleryItemsFromPaths([
    "/f/home/test/IMG_2062.HEIC",
    "/f/home/test/IMG_2062.MOV",
  ]),
): void {
  vi.stubGlobal("ResizeObserver", GalleryResizeObserver);
  vi.stubGlobal("IS_REACT_ACT_ENVIRONMENT", true);
  Object.defineProperty(HTMLElement.prototype, "scrollIntoView", {
    configurable: true,
    value: vi.fn(),
  });
  vi.spyOn(HTMLElement.prototype, "clientWidth", "get").mockReturnValue(960);
  vi.spyOn(HTMLElement.prototype, "clientHeight", "get").mockReturnValue(650);
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
  vi.spyOn(HTMLMediaElement.prototype, "pause").mockImplementation(() => {});
  vi.stubGlobal(
    "fetch",
    vi.fn(async (url: string) =>
      url.endsWith(".HEIC")
        ? mediaRange(`HEIC metadata ${identifier}`, false)
        : mediaRange(`com.apple.quicktime.content.identifier ${identifier}`, true),
    ),
  );
  render(
    <StreamProvider>
      <ThemeProvider theme={createAppTheme("default", "dark")}>
        <Gallery items={items} layout="mosaic" />
      </ThemeProvider>
    </StreamProvider>,
  );
}

/** Opens the verified still and starts its associated movie from the viewer. */
async function openLivePhotoMotion(): Promise<HTMLVideoElement> {
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
  renderLivePhotoGallery();
  fireEvent.click(await screen.findByRole("button", { name: "Open IMG_2062.HEIC" }));
  const dialog = screen.getByRole("dialog");
  await act(async () => {
    fireEvent.load(within(dialog).getByAltText("IMG_2062.HEIC"));
  });
  fireEvent.click(within(dialog).getByRole("button", { name: "Play Live Photo IMG_2062.HEIC" }));
  const video = dialog.querySelector("video");
  if (!video) throw new Error("Live Photo motion element was not rendered");
  return video;
}

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
  vi.restoreAllMocks();
  Reflect.deleteProperty(HTMLElement.prototype, "scrollIntoView");
});

// Requirements: moments/FR-003. Matching still and movie originals render as one gallery item.
test("verified HEIC and MOV companions render as one Live Photo", async () => {
  renderLivePhotoGallery();
  await waitFor(() => expect(screen.getByText("1 media items")).toBeTruthy());
});

// Requirements: moments/FR-003. The composed item offers its motion control after loading.
test("verified Live Photo offers motion playback in the viewer", async () => {
  renderLivePhotoGallery();
  fireEvent.click(await screen.findByRole("button", { name: "Open IMG_2062.HEIC" }));
  const dialog = screen.getByRole("dialog");
  await act(async () => {
    fireEvent.load(within(dialog).getByAltText("IMG_2062.HEIC"));
  });
  expect(
    within(dialog).getByRole("button", { name: "Play Live Photo IMG_2062.HEIC" }),
  ).toBeTruthy();
});

// Requirements: moments/FR-003. The Live Photo motion control uses its verified MOV.
test("Live Photo playback starts the associated movie", async () => {
  const video = await openLivePhotoMotion();
  expect(video.getAttribute("src")).toBe("/s/home/test/IMG_2062.MOV");
});

// Requirements: moments/FR-003. Failed HEIC display leaves the paired movie visible on the wall.
test("a Live Photo falls back to its movie when the still cannot display", async () => {
  renderLivePhotoGallery();
  fireEvent.error(await screen.findByAltText("IMG_2062.HEIC"));
  expect(await screen.findByRole("button", { name: "Play video IMG_2062.MOV" })).toBeTruthy();
});

// Requirements: moments/FR-003. The viewer uses the paired movie after HEIC display fails.
test("the viewer falls back to the paired movie when the still cannot display", async () => {
  renderLivePhotoGallery();
  fireEvent.click(await screen.findByRole("button", { name: "Open IMG_2062.HEIC" }));
  const dialog = screen.getByRole("dialog");
  fireEvent.error(within(dialog).getByAltText("IMG_2062.HEIC"));
  await waitFor(() =>
    expect(dialog.querySelector("video")?.getAttribute("src")).toBe("/s/home/test/IMG_2062.MOV"),
  );
});

// Requirements: moments/FR-003. A fallback Live Photo movie waits for explicit viewer playback.
test("the viewer does not autoplay Live Photo motion after its still fails", async () => {
  const play = vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
  renderLivePhotoGallery();
  fireEvent.click(await screen.findByRole("button", { name: "Open IMG_2062.HEIC" }));
  const dialog = screen.getByRole("dialog");
  fireEvent.error(within(dialog).getByAltText("IMG_2062.HEIC"));
  await waitFor(() =>
    expect(dialog.querySelector("video")?.getAttribute("src")).toBe("/s/home/test/IMG_2062.MOV"),
  );
  expect(play).not.toHaveBeenCalled();
});

// Requirements: moments/FR-003. A failed paired movie uses the transcoded stream.
test("the viewer transcodes a paired movie when its original cannot play", async () => {
  renderLivePhotoGallery();
  fireEvent.click(await screen.findByRole("button", { name: "Open IMG_2062.HEIC" }));
  const dialog = screen.getByRole("dialog");
  fireEvent.error(within(dialog).getByAltText("IMG_2062.HEIC"));
  const video = dialog.querySelector("video");
  if (!video) throw new Error("Paired movie was not rendered");
  fireEvent.error(video);
  await waitFor(() =>
    expect(video.getAttribute("src")).toBe("/s/transcode/home/test/IMG_2062.MOV?start=0"),
  );
});

// Requirements: moments/FR-003. A standalone HEIC remains an image on display failure.
test("a standalone HEIC remains an image when its preview fails", async () => {
  renderLivePhotoGallery(galleryItemsFromPaths(["/f/home/test/solo.HEIC"]));
  fireEvent.error(await screen.findByAltText("solo.HEIC"));
  expect(await screen.findByText("Image unavailable")).toBeTruthy();
});

// Requirements: moments/FR-003. Motion failure retains the still and explains the error.
test("Live Photo reports failed original and fallback motion", async () => {
  const video = await openLivePhotoMotion();
  fireEvent.error(video);
  fireEvent.error(video);
  expect(await screen.findByText("Live Photo motion is unavailable.")).toBeTruthy();
});

// Requirements: moments/FR-003. A finished Live Photo returns to its still.
test("Live Photo returns to the still after motion ends", async () => {
  const video = await openLivePhotoMotion();
  fireEvent.ended(video);
  expect(screen.getByRole("button", { name: "Play Live Photo IMG_2062.HEIC" })).toBeTruthy();
});

// Requirements: moments/FR-003. The viewer exposes motion buffering to the user.
test("Live Photo shows a spinner while motion buffers", async () => {
  const video = await openLivePhotoMotion();
  fireEvent.waiting(video);
  expect(screen.getByLabelText("Buffering video")).toBeTruthy();
});

// Requirements: moments/FR-003. Playing motion clears a prior buffering state.
test("Live Photo clears buffering when playback resumes", async () => {
  const video = await openLivePhotoMotion();
  fireEvent.waiting(video);
  fireEvent.playing(video);
  expect(screen.queryByLabelText("Buffering video")).toBeNull();
});

// Requirements: moments/FR-003. A playable motion frame clears a prior wait.
test("Live Photo clears buffering when frames become ready", async () => {
  const video = await openLivePhotoMotion();
  fireEvent.waiting(video);
  fireEvent.canPlay(video);
  expect(screen.queryByLabelText("Buffering video")).toBeNull();
});

// Requirements: moments/FR-003. A standalone movie starts from its authorized original.
test("standalone movie playback starts only after the user presses Play", async () => {
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
  renderLivePhotoGallery(galleryItemsFromPaths(["/f/home/test/clip.MOV"]));
  fireEvent.click(await screen.findByRole("button", { name: "Play video clip.MOV" }));
  expect(screen.getByLabelText("Play video clip.MOV").getAttribute("src")).toBe(
    "/s/home/test/clip.MOV",
  );
});

// Requirements: moments/FR-003. A failed original movie can switch to transcoding.
test("standalone movie falls back to a transcoded stream", async () => {
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
  renderLivePhotoGallery(galleryItemsFromPaths(["/f/home/test/clip.MOV"]));
  fireEvent.click(await screen.findByRole("button", { name: "Play video clip.MOV" }));
  fireEvent.error(screen.getByLabelText("Play video clip.MOV"));
  await waitFor(() =>
    expect(screen.getByLabelText("Play video clip.MOV").getAttribute("src")).toBe(
      "/s/transcode/home/test/clip.MOV?start=0",
    ),
  );
});

// Requirements: moments/FR-003. Repeated playback failure stays visible to the user.
test("standalone movie reports when its fallback also fails", async () => {
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
  renderLivePhotoGallery(galleryItemsFromPaths(["/f/home/test/clip.MOV"]));
  fireEvent.click(await screen.findByRole("button", { name: "Play video clip.MOV" }));
  fireEvent.error(screen.getByLabelText("Play video clip.MOV"));
  fireEvent.error(screen.getByLabelText("Play video clip.MOV"));
  expect(await screen.findByText("Video unavailable")).toBeTruthy();
});

// Requirements: moments/FR-003. A fallback movie can seek by starting a fresh stream.
test("transcoded movie can seek forward in the wall", async () => {
  vi.spyOn(HTMLMediaElement.prototype, "play").mockResolvedValue(undefined);
  renderLivePhotoGallery(galleryItemsFromPaths(["/f/home/test/clip.MOV"]));
  fireEvent.click(await screen.findByRole("button", { name: "Play video clip.MOV" }));
  fireEvent.error(screen.getByLabelText("Play video clip.MOV"));
  fireEvent.click(screen.getByRole("button", { name: "Forward 30s" }));
  await waitFor(() =>
    expect(screen.getByLabelText("Play video clip.MOV").getAttribute("src")).toBe(
      "/s/transcode/home/test/clip.MOV?start=30",
    ),
  );
});

// Requirements: moments/FR-003. Visitors see the saved layout without a local override.
test("gallery wall has no visitor layout selector", async () => {
  renderLivePhotoGallery(galleryItemsFromPaths(["/f/home/test/photo.jpg"]));
  await screen.findByRole("button", { name: "Open photo.jpg" });
  expect(screen.queryByRole("combobox", { name: "Layout" })).toBeNull();
});

// Requirements: moments/FR-003. Viewer navigation changes the selected original.
test("viewer navigates between separate photos", async () => {
  renderLivePhotoGallery(
    galleryItemsFromPaths(["/f/home/test/first.jpg", "/f/home/test/second.jpg"]),
  );
  fireEvent.click(await screen.findByRole("button", { name: "Open first.jpg" }));
  fireEvent.click(screen.getByRole("button", { name: "Next item" }));
  expect(screen.getByText("2 / 2")).toBeTruthy();
});

// Requirements: moments/FR-003. Closing the viewer returns focus to the wall.
test("viewer closes without removing the gallery wall", async () => {
  renderLivePhotoGallery(galleryItemsFromPaths(["/f/home/test/photo.jpg"]));
  fireEvent.click(await screen.findByRole("button", { name: "Open photo.jpg" }));
  fireEvent.click(screen.getByRole("button", { name: "Close viewer" }));
  await waitFor(() => expect(screen.queryByRole("dialog")).toBeNull());
});
