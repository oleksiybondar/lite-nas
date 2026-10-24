import Alert from "@mui/material/Alert";
import Box from "@mui/material/Box";
import CircularProgress from "@mui/material/CircularProgress";
import type { ReactElement } from "react";
import { useEffect, useState } from "react";
import { Gallery } from "./Gallery";
import type { GalleryLayoutMode } from "./GalleryLayout";
import { type GalleryItem, galleryFixtureSchema, galleryItemsFromPaths } from "./gallery-model";

type DemoContent = { items: GalleryItem[]; defaultLayout: GalleryLayoutMode };

/** Loads the bundled demo list while leaving gallery presentation independent of its source. */
export const DemoGallery = (): ReactElement => {
  const [content, setContent] = useState<DemoContent | null>(null);
  const [error, setError] = useState(false);
  useEffect(() => {
    const controller = new AbortController();
    void fetch(`${import.meta.env.BASE_URL}dev-fixtures/gallery.json`, {
      signal: controller.signal,
    })
      .then(async (response) => {
        if (!response.ok) throw new Error("Demo fixture unavailable");
        const value: unknown = await response.json();
        const fixture = galleryFixtureSchema.parse(value);
        return {
          items: galleryItemsFromPaths(fixture.items),
          defaultLayout: fixture.defaultLayout,
        };
      })
      .then(setContent)
      .catch(() => {
        if (!controller.signal.aborted) setError(true);
      });
    return () => controller.abort();
  }, []);
  if (error) return <Alert severity="error">The demo gallery could not be loaded.</Alert>;
  if (!content)
    return (
      <Box display="flex" justifyContent="center" py={8}>
        <CircularProgress aria-label="Loading gallery" />
      </Box>
    );
  return <Gallery items={content.items} defaultLayout={content.defaultLayout} />;
};
