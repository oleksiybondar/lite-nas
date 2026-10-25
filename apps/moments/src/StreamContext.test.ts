import { expect, test } from "vitest";
import { streamProtocol } from "./StreamContext";

test("original media keeps the host path on the Stream Service mount", () => {
  expect(streamProtocol.original("/f/home/test/movie.avi")).toBe("/s/home/test/movie.avi");
});

test("transcoded media includes a validated time seek", () => {
  expect(streamProtocol.transcode("/f/home/test/movie.mov", 30)).toBe(
    "/s/transcode/home/test/movie.mov?start=30",
  );
});

test("transcoding rejects an invalid time seek", () => {
  expect(streamProtocol.transcode("/f/home/test/movie.mov", -1)).toBeNull();
});

test("HEIC stills use the authorized JPEG preview route", () => {
  expect(streamProtocol.image("/f/home/test/IMG_1864.HEIC")).toBe(
    "/s/preview/home/test/IMG_1864.HEIC",
  );
});

test("JPEG images retain their original File Service URL", () => {
  expect(streamProtocol.image("/f/home/test/photo.jpeg")).toBe("/f/home/test/photo.jpeg");
});

test("protocol rejects an untrusted URL", () => {
  expect(streamProtocol.original("https://example.com/movie.mov")).toBeNull();
});
