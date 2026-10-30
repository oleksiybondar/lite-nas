// @vitest-environment jsdom

import { cleanup, screen } from "@testing-library/react";
import { afterEach, expect, test } from "vitest";
import { GalleryUploaderLabel } from "./GalleryUploaderLabel";
import { renderWithLanguage } from "./test-utils/language";

afterEach(cleanup);

test("uploader label names the member from the folder path", () => {
  renderWithLanguage(<GalleryUploaderLabel uploaderId="1002" />);
  expect(screen.getByText("Uploaded by 1002")).toBeTruthy();
});

test("uploader label includes a user icon", () => {
  renderWithLanguage(<GalleryUploaderLabel uploaderId="1002" />);
  expect(screen.getByTestId("uploader-icon")).toBeTruthy();
});
