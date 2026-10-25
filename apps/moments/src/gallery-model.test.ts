import { readFileSync } from "node:fs";
import { expect, test } from "vitest";
import {
  galleryDownloadUrl,
  galleryFixtureSchema,
  galleryIsHeic,
  galleryItemsFromPaths,
  imageDelaySchema,
} from "./gallery-model";

test("fixture originals are classified in their existing order", () => {
  const paths = ["/f/home/test/photo.jpeg", "/f/home/test/movie.mov"];
  const fixture = galleryFixtureSchema.parse({ defaultLayout: "masonry", items: paths });
  expect(galleryItemsFromPaths(fixture.items).map((item) => item.kind)).toEqual(["image", "video"]);
});

test.each([
  "https://example.com/photo.jpg",
  "/f/home/test/../secret.jpg",
  "/f/home/test/%GG.jpg",
])("unsafe fixture URL is rejected: %s", (path) => {
  expect(galleryFixtureSchema.safeParse({ defaultLayout: "mosaic", items: [path] }).success).toBe(
    false,
  );
});

test("fixture rejects an unknown default layout", () => {
  expect(galleryFixtureSchema.safeParse({ defaultLayout: "random", items: [] }).success).toBe(
    false,
  );
});

test("bundled demo fixture matches the validated layout and item format", () => {
  const fixture: unknown = JSON.parse(
    readFileSync(new URL("../public/dev-fixtures/gallery.json", import.meta.url), "utf8"),
  );
  expect(galleryFixtureSchema.safeParse(fixture).success).toBe(true);
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
