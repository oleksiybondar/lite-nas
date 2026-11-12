// @vitest-environment jsdom

import { act, cleanup, fireEvent, render, screen } from "@testing-library/react";
import type { ReactElement } from "react";
import { afterEach, expect, test, vi } from "vitest";
import { GalleryProvider, useGalleryMediaUi } from "./GalleryContext";

/** Exposes reducer state and deterministic actions for provider-level tests. */
const MediaUiProbe = (): ReactElement => {
  const { state, dispatch } = useGalleryMediaUi();
  return (
    <>
      <output data-testid="media-state">{JSON.stringify(state)}</output>
      <button
        type="button"
        onClick={() => dispatch({ type: "image/ready", itemId: "first", url: "/first.jpg" })}
      >
        Ready first
      </button>
      <button
        type="button"
        onClick={() => dispatch({ type: "image/ready", itemId: "second", url: "/second.jpg" })}
      >
        Ready second
      </button>
      <button
        type="button"
        onClick={() =>
          dispatch({
            type: "wall/windowChanged",
            rendered: ["first", "second"],
            visible: ["first", "second"],
            direction: "down",
          })
        }
      >
        Keep tiles
      </button>
      <button
        type="button"
        onClick={() =>
          dispatch({
            type: "wall/windowChanged",
            rendered: [],
            visible: [],
            direction: "down",
          })
        }
      >
        Clear wall
      </button>
    </>
  );
};

/** Mounts the media UI reducer with an empty gallery collection. */
const renderProbe = (): void => {
  render(
    <GalleryProvider items={[]}>
      <MediaUiProbe />
    </GalleryProvider>,
  );
};

afterEach(() => {
  cleanup();
  vi.useRealTimers();
});

test("image readiness publishes the first update immediately", () => {
  renderProbe();
  fireEvent.click(screen.getByRole("button", { name: "Ready first" }));

  expect(screen.getByTestId("media-state").textContent).toContain(
    '"first":{"url":"/first.jpg","status":"ready"}',
  );
});

test("image readiness buffers updates while the cooldown is active", () => {
  vi.useFakeTimers();
  renderProbe();
  fireEvent.click(screen.getByRole("button", { name: "Keep tiles" }));
  fireEvent.click(screen.getByRole("button", { name: "Ready first" }));
  act(() => vi.advanceTimersByTime(50));
  fireEvent.click(screen.getByRole("button", { name: "Ready second" }));
  act(() => vi.advanceTimersByTime(49));

  expect(screen.getByTestId("media-state").textContent).not.toContain('"second":');
});

test("image readiness batches later updates at the fixed 100 ms deadline", () => {
  vi.useFakeTimers();
  renderProbe();
  fireEvent.click(screen.getByRole("button", { name: "Keep tiles" }));
  fireEvent.click(screen.getByRole("button", { name: "Ready first" }));
  act(() => vi.advanceTimersByTime(50));
  fireEvent.click(screen.getByRole("button", { name: "Ready second" }));
  act(() => vi.advanceTimersByTime(50));

  expect(screen.getByTestId("media-state").textContent).toContain(
    '"second":{"url":"/second.jpg","status":"ready"}',
  );
});

test("wall window changes prune media records that are no longer retained", () => {
  renderProbe();
  fireEvent.click(screen.getByRole("button", { name: "Ready first" }));
  fireEvent.click(screen.getByRole("button", { name: "Clear wall" }));

  expect(screen.getByTestId("media-state").textContent).toContain('"items":{}');
});
