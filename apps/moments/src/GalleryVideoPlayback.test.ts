import { expect, test, vi } from "vitest";
import { attemptVideoAutoplay } from "./GalleryVideoPlayback";

test("video autoplay starts when the browser permits playback", async () => {
  const video = { play: vi.fn().mockResolvedValue(undefined) };
  await expect(attemptVideoAutoplay(video, true)).resolves.toBe("started");
});

test("autoplay rejection remains a manual-play state", async () => {
  const video = { play: vi.fn().mockRejectedValue(new DOMException("blocked", "NotAllowedError")) };
  await expect(attemptVideoAutoplay(video, true)).resolves.toBe("blocked");
});

test("disabling video autoplay leaves the original paused", async () => {
  const video = { play: vi.fn() };
  await expect(attemptVideoAutoplay(video, false)).resolves.toBe("disabled");
});

test("unsupported video playback is treated as a media failure", async () => {
  const video = {
    play: vi.fn().mockRejectedValue(new DOMException("unsupported", "NotSupportedError")),
  };
  await expect(attemptVideoAutoplay(video, true)).resolves.toBe("failed");
});

test("interrupted playback is not treated as an autoplay block", async () => {
  const video = { play: vi.fn().mockRejectedValue(new DOMException("interrupted", "AbortError")) };
  await expect(attemptVideoAutoplay(video, true)).resolves.toBe("interrupted");
});

test("disabled video autoplay does not invoke native playback", async () => {
  const video = { play: vi.fn() };
  await attemptVideoAutoplay(video, false);
  expect(video.play).not.toHaveBeenCalled();
});
