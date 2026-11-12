import { expect, test } from "vitest";
import { GalleryImageQueue } from "./GalleryImageQueue";

test("image queue grants ten original download slots by default", () => {
  const queue = new GalleryImageQueue();
  const urls = Array.from({ length: 12 }, (_, index) => `/f/home/test/${index}.jpg`);
  queue.want(urls);
  expect(urls.filter((url) => queue.status(url) === "loading")).toHaveLength(10);
});

test("image queue uses the configured parallel download limit", () => {
  const queue = new GalleryImageQueue(3);
  queue.want(Array.from({ length: 5 }, (_, index) => `/f/home/test/${index}.jpg`));
  expect(
    [0, 1, 2, 3, 4].filter((index) => queue.status(`/f/home/test/${index}.jpg`) === "loading"),
  ).toHaveLength(3);
});

test("a settled image grants its slot to the next needed original", () => {
  const queue = new GalleryImageQueue();
  const urls = Array.from({ length: 4 }, (_, index) => `/f/home/test/${index}.jpg`);
  queue.want(urls);
  queue.complete(urls[0], true);
  expect(queue.status(urls[2])).toBe("loading");
});

test("releasing an offscreen image removes its ready state", () => {
  const queue = new GalleryImageQueue();
  queue.want(["/f/home/test/one.jpg"]);
  queue.complete("/f/home/test/one.jpg", true);
  queue.retain([]);
  expect(queue.status("/f/home/test/one.jpg")).toBe("queued");
});

test("a canceled image error does not poison the file's state", () => {
  const queue = new GalleryImageQueue();
  const url = "/f/home/test/photo.jpeg";
  queue.want([url]);
  queue.want([]);
  queue.fail(url);
  expect(queue.status(url)).toBe("queued");
});

test("a failed image can be retried in the active viewer", () => {
  const queue = new GalleryImageQueue();
  const url = "/f/home/test/photo.jpeg";
  queue.want([url]);
  queue.fail(url);
  queue.retry(url);
  expect(queue.status(url)).toBe("loading");
});

test("an old viewer error cannot invalidate an image retained on the wall", () => {
  const queue = new GalleryImageQueue();
  const url = "/f/home/test/photo.jpeg";
  queue.want([url]);
  queue.complete(url, true);
  queue.want([]);
  queue.fail(url);
  expect(queue.status(url)).toBe("ready");
});

test("late failure from a canceled attempt does not fail a restarted request", () => {
  const queue = new GalleryImageQueue();
  const url = "/f/home/test/photo.jpeg";
  queue.want([url]);
  const oldAttempt = queue.attempt(url);
  queue.want([]);
  queue.want([url]);
  queue.fail(url, oldAttempt);
  expect(queue.status(url)).toBe("loading");
});
