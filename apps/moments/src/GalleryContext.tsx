import type { ReactElement, ReactNode } from "react";
import { createContext, useContext, useEffect, useMemo } from "react";
import { GalleryImageQueue } from "./GalleryImageQueue";
import { GalleryMetadataReader } from "./GalleryMetadataReader";
import type { GalleryItem } from "./gallery-model";

/** Stable collection metadata and one image request coordinator shared by both views. */
export type GalleryState = {
  items: readonly GalleryItem[];
  queue: GalleryImageQueue;
  metadata: GalleryMetadataReader;
};
const GalleryContext = createContext<GalleryState | null>(null);

/** Owns the collection and releases shared requests when the gallery unmounts. */
export const GalleryProvider = ({
  items,
  children,
}: {
  items: readonly GalleryItem[];
  children: ReactNode;
}): ReactElement => {
  const queue = useMemo(() => new GalleryImageQueue(), []);
  const metadata = useMemo(() => new GalleryMetadataReader(), []);
  const value = useMemo(() => ({ items, queue, metadata }), [items, queue, metadata]);
  useEffect(
    () => () => {
      queue.dispose();
      metadata.dispose();
    },
    [queue, metadata],
  );
  return <GalleryContext.Provider value={value}>{children}</GalleryContext.Provider>;
};

/** Reads validated gallery metadata and its shared request coordinator. */
export const useGallery = (): GalleryState => {
  const value = useContext(GalleryContext);
  if (!value) throw new Error("GalleryProvider is required");
  return value;
};
