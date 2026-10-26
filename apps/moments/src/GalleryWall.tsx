// biome-ignore-all lint/complexity/noExcessiveLinesPerFunction: wall request coordination and virtualized rendering share one viewport state.
import Box from "@mui/material/Box";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useEffect, useSyncExternalStore } from "react";
import { useCarousel } from "./CarouselContext";
import { useGallery } from "./GalleryContext";
import { GalleryWallItem } from "./GalleryWallItem";
import { wallImageRequests } from "./GalleryWindow";
import { useWall } from "./WallContext";

/** Materializes a small wall window and coordinates viewport-only image requests. */
export const GalleryWall = (): ReactElement => {
  const { items, queue } = useGallery();
  const { wallRef, onScroll, layoutHeight, window: wallWindow, scrolling } = useWall();
  const { selected, current, open } = useCarousel();
  const imageVersion = useSyncExternalStore(queue.subscribe, queue.getSnapshot);
  useEffect(() => {
    const retained = wallWindow.rendered
      .filter((tile) => items[tile.index].kind === "image")
      .map((tile) => items[tile.index].url);
    if (current?.kind === "image") retained.push(current.url);
    queue.retain(retained);
  }, [current, items, queue, wallWindow.rendered]);
  // Queue snapshots recalculate requests after image slots complete.
  // biome-ignore lint/correctness/useExhaustiveDependencies: queue status is read inside the request planner callback.
  useEffect(() => {
    if (current?.kind === "image") {
      queue.want([current.url]);
      return;
    }
    if (selected !== null || scrolling) {
      queue.want([]);
      return;
    }
    queue.want(wallImageRequests(items, wallWindow.visible, queue.status, queue.maxConcurrent));
  }, [current, imageVersion, items, queue, scrolling, selected, wallWindow.visible]);
  return (
    <Box data-testid="moments-gallery" sx={{ width: "100%" }}>
      <Stack
        direction={{ xs: "column", sm: "row" }}
        alignItems={{ xs: "start", sm: "center" }}
        justifyContent="space-between"
        gap={2}
        mb={2}
      >
        <Box>
          <Typography variant="h5">Media</Typography>
          <Typography color="text.secondary" variant="body2">
            {items.length} media items
          </Typography>
        </Box>
      </Stack>
      <Box
        ref={wallRef}
        onScroll={onScroll}
        sx={{
          height: "min(70dvh, 850px)",
          minHeight: 360,
          overflowY: "auto",
          overflowX: "hidden",
          borderRadius: 2,
          bgcolor: "action.hover",
          position: "relative",
        }}
      >
        <Box sx={{ height: layoutHeight, position: "relative" }}>
          {wallWindow.rendered.map((tile) => (
            <GalleryWallItem
              key={items[tile.index].id}
              item={items[tile.index]}
              tile={tile}
              onOpen={() => open(tile.index)}
            />
          ))}
        </Box>
      </Box>
    </Box>
  );
};
