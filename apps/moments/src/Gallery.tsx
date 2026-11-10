import type { ReactElement } from "react";
import { CarouselProvider } from "./CarouselContext";
import { GalleryProvider, type GalleryState } from "./GalleryContext";
import type { GalleryLayoutMode } from "./GalleryLayout";
import { GalleryViewer } from "./GalleryViewer";
import { GalleryWall } from "./GalleryWall";
import type { GalleryItem } from "./gallery-model";
import { WallProvider } from "./WallContext";

/** Composes the reusable gallery's collection, wall, and viewer state owners. */
export const Gallery = ({
  items,
  layout,
  collectionActions,
  maxConcurrent,
  imageUiCooldownMs,
}: {
  items: readonly GalleryItem[];
  layout: GalleryLayoutMode;
  collectionActions?: GalleryState["collectionActions"];
  maxConcurrent?: number;
  imageUiCooldownMs?: number;
}): ReactElement => (
  <GalleryProvider
    items={items}
    collectionActions={collectionActions}
    maxConcurrent={maxConcurrent}
    imageUiCooldownMs={imageUiCooldownMs}
  >
    <WallProvider layout={layout}>
      <CarouselProvider>
        <GalleryWall />
        <GalleryViewer />
      </CarouselProvider>
    </WallProvider>
  </GalleryProvider>
);
