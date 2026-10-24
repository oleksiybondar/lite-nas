import type { ReactElement } from "react";
import { CarouselProvider } from "./CarouselContext";
import { GalleryProvider } from "./GalleryContext";
import type { GalleryLayoutMode } from "./GalleryLayout";
import { GalleryViewer } from "./GalleryViewer";
import { GalleryWall } from "./GalleryWall";
import type { GalleryItem } from "./gallery-model";
import { WallProvider } from "./WallContext";

/** Composes the reusable gallery's collection, wall, and viewer state owners. */
export const Gallery = ({
  items,
  defaultLayout,
}: {
  items: readonly GalleryItem[];
  defaultLayout: GalleryLayoutMode;
}): ReactElement => (
  <GalleryProvider items={items}>
    <WallProvider defaultLayout={defaultLayout}>
      <CarouselProvider>
        <GalleryWall />
        <GalleryViewer />
      </CarouselProvider>
    </WallProvider>
  </GalleryProvider>
);
