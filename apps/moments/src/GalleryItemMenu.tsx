import MoreVertIcon from "@mui/icons-material/MoreVert";
import IconButton from "@mui/material/IconButton";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import type { ReactElement } from "react";
import { useId, useState } from "react";
import type { GalleryItem } from "./gallery-model";

/** Gives wall and carousel items the same menu surface for future collection actions. */
export const GalleryItemMenu = ({
  item,
  wall = false,
}: {
  item: GalleryItem;
  wall?: boolean;
}): ReactElement => {
  const menuId = useId();
  const [anchor, setAnchor] = useState<HTMLElement | null>(null);
  const open = anchor !== null;
  return (
    <>
      <IconButton
        size={wall ? "small" : "medium"}
        color="inherit"
        aria-label={`More options for ${item.filename}`}
        aria-haspopup="menu"
        aria-controls={open ? menuId : undefined}
        aria-expanded={open ? "true" : undefined}
        onClick={(event) => setAnchor(event.currentTarget)}
        sx={
          wall
            ? {
                color: "common.white",
                bgcolor: "rgba(0,0,0,0.55)",
                "&:hover": { bgcolor: "rgba(0,0,0,0.8)" },
              }
            : undefined
        }
      >
        <MoreVertIcon fontSize={wall ? "small" : "medium"} />
      </IconButton>
      <Menu
        id={menuId}
        anchorEl={anchor}
        open={open}
        onClose={() => setAnchor(null)}
        anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
        transformOrigin={{ vertical: "top", horizontal: "right" }}
      >
        <MenuItem disabled>Add to collection</MenuItem>
      </Menu>
    </>
  );
};
