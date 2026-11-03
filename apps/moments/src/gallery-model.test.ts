import { expect, test } from "vitest";
import {
  galleryDownloadUrl,
  galleryFileUrlSchema,
  galleryIsHeic,
  galleryItemsFromPaths,
  galleryUploaderFromPath,
  imageDelaySchema,
} from "./gallery-model";

test("originals are classified in their existing order", () => {
  const paths = ["/f/home/test/photo.jpeg", "/f/home/test/movie.mov"];
  expect(galleryItemsFromPaths(paths).map((item) => item.kind)).toEqual(["image", "video"]);
});

test.each([
  "https://example.com/photo.jpg",
  "/f/home/test/../secret.jpg",
  "/f/home/test/%GG.jpg",
])("unsafe original URL is rejected: %s", (path) => {
  expect(galleryFileUrlSchema.safeParse(path).success).toBe(false);
});

test.each([1000, 3000, 5000])("slideshow accepts the %i ms image delay", (delay) => {
  expect(imageDelaySchema.safeParse(delay).success).toBe(true);
});

test("slideshow rejects an unsupported image delay", () => {
  expect(imageDelaySchema.safeParse(2000).success).toBe(false);
});

test("download uses the attachment route for the same original path", () => {
  expect(galleryDownloadUrl("/f/home/test/photo.jpeg")).toBe("/d/home/test/photo.jpeg");
});

test("download rejects a URL outside the original File Service route", () => {
  expect(galleryDownloadUrl("https://example.com/photo.jpeg")).toBeNull();
});

test("HEIC items are images in the gallery", () => {
  expect(galleryItemsFromPaths(["/f/home/test/IMG_2063.HEIC"])[0].kind).toBe("image");
});

test("HEIC originals are recognized for JPEG preview routing", () => {
  expect(galleryIsHeic("/f/home/test/IMG_2063.HEIC")).toBe(true);
});

test("an immediate album member folder identifies the uploader", () => {
  expect(galleryUploaderFromPath("/home/Photos/1002/photo.jpg", "/home/Photos")).toBe("1002");
});

test.each([
  ["/home/Photos/photo.jpg", "/home/Photos"],
  ["/home/Photos/trip/2026/photo.jpg", "/home/Photos"],
  ["/home/PhotosElse/1002/photo.jpg", "/home/Photos"],
])("non-member upload path %s has no uploader attribution", (path, albumDirectory) => {
  expect(galleryUploaderFromPath(path, albumDirectory)).toBeNull();
});

test("current user's own upload folder has no other-member attribution", () => {
  expect(galleryUploaderFromPath("/home/Photos/1001/photo.jpg", "/home/Photos", "1001")).toBeNull();
});
