// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: resize, scroll, and layout state form the wall provider lifecycle.
import type { ReactElement, ReactNode, RefObject, UIEvent } from "react";
import { createContext, useContext, useEffect, useMemo, useRef, useState } from "react";
import { useGallery } from "./GalleryContext";
import { type GalleryLayoutMode, planGallery } from "./GalleryLayout";
import { type GalleryWindow, galleryWindow } from "./GalleryWindow";

/** Wall-only geometry, viewport, and scroll activity. */
export type WallState = {
  wallRef: RefObject<HTMLDivElement>;
  scrolling: boolean;
  activeVideoId: string | null;
  setActiveVideoId: (id: string | null) => void;
  layoutHeight: number;
  window: GalleryWindow;
  onScroll: (event: UIEvent<HTMLDivElement>) => void;
};
const WallContext = createContext<WallState | null>(null);

/** Keeps scroll position and layout mounted while the viewer is open. */
export const WallProvider = ({
  layout,
  children,
}: {
  layout: GalleryLayoutMode;
  children: ReactNode;
}): ReactElement => {
  const { items } = useGallery();
  const wallRef = useRef<HTMLDivElement>(null);
  const [width, setWidth] = useState(0);
  const [viewportHeight, setViewportHeight] = useState(650);
  const [scrollTop, setScrollTop] = useState(0);
  const [scrolling, setScrolling] = useState(false);
  const [activeVideoId, setActiveVideoId] = useState<string | null>(null);
  const previousLayout = useRef(layout);
  useEffect(() => {
    const element = wallRef.current;
    if (!element) return;
    const observer = new ResizeObserver(() => {
      setWidth(element.clientWidth);
      setViewportHeight(element.clientHeight);
    });
    observer.observe(element);
    return () => observer.disconnect();
  }, []);
  useEffect(() => {
    if (!scrolling) return;
    const timer = window.setTimeout(() => setScrolling(false), 120);
    return () => window.clearTimeout(timer);
  }, [scrolling]);
  const plan = useMemo(() => planGallery(items, width, layout), [items, width, layout]);
  const wallWindow = useMemo(
    () => galleryWindow(plan.tiles, scrollTop, viewportHeight),
    [plan, scrollTop, viewportHeight],
  );
  /** Returns to the start when the saved gallery layout changes. */
  useEffect(() => {
    if (previousLayout.current === layout) return;
    previousLayout.current = layout;
    if (wallRef.current) wallRef.current.scrollTop = 0;
    setScrollTop(0);
  }, [layout]);
  /** Updates placeholders immediately and resumes original requests after scrolling settles. */
  const onScroll = (event: UIEvent<HTMLDivElement>): void => {
    setScrollTop(event.currentTarget.scrollTop);
    setScrolling(true);
  };
  const value = {
    wallRef,
    scrolling,
    activeVideoId,
    setActiveVideoId,
    layoutHeight: plan.height,
    window: wallWindow,
    onScroll,
  };
  return <WallContext.Provider value={value}>{children}</WallContext.Provider>;
};

/** Reads wall-only state without giving the viewer ownership of scrolling. */
export const useWall = (): WallState => {
  const value = useContext(WallContext);
  if (!value) throw new Error("WallProvider is required");
  return value;
};
