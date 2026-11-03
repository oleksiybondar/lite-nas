import type { ReactElement, ReactNode } from "react";
import { createContext, useContext, useEffect, useMemo } from "react";
import { GalleryImageQueue } from "./GalleryImageQueue";
import { GalleryMetadataReader } from "./GalleryMetadataReader";
import type { GalleryItem } from "./gallery-model";
import { composeLivePhotoItems } from "./LivePhotoMetadata";
import type { GalleryRecord } from "./schemas";

/** Stable collection metadata and one image request coordinator shared by both views. */
export type GalleryState = {
  items: readonly GalleryItem[];
  queue: GalleryImageQueue;
  metadata: GalleryMetadataReader;
  collectionActions?: {
    collections: readonly GalleryRecord[];
    canRemove: (path: string) => boolean;
    add: (collectionId: string, path: string) => Promise<void>;
    remove: (path: string) => Promise<void>;
    createAndAdd: (name: string, path: string) => Promise<void>;
  };
};
const GalleryContext = createContext<GalleryState | null>(null);

/** Owns the collection and releases shared requests when the gallery unmounts. */
export const GalleryProvider = ({
  items,
  children,
  collectionActions,
  maxConcurrent = 10,
}: {
  items: readonly GalleryItem[];
  children: ReactNode;
  collectionActions?: GalleryState["collectionActions"];
  maxConcurrent?: number;
}): ReactElement => {
  const queue = useMemo(() => new GalleryImageQueue(maxConcurrent), [maxConcurrent]);
  const metadata = useMemo(() => new GalleryMetadataReader(), []);
  const composedItems = useMemo(() => composeLivePhotoItems(items), [items]);
  const value = useMemo(
    () => ({ items: composedItems, queue, metadata, collectionActions }),
    [composedItems, queue, metadata, collectionActions],
  );
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
