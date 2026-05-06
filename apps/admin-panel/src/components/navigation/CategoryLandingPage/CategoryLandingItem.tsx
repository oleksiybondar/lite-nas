import Box from "@mui/material/Box";
import Grid from "@mui/material/Grid";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { Link as RouterLink } from "react-router-dom";
import type { CategoryLandingCard } from "./types";

type CategoryLandingItemProps = {
  /**
   * Card data rendered as a clickable category entry.
   */
  card: CategoryLandingCard;
};

/**
 * Clickable category item used by landing pages to expose child routes.
 */
export const CategoryLandingItem = ({ card }: CategoryLandingItemProps): ReactElement => {
  return (
    <Grid size={{ md: 4, sm: 6, xs: 12 }}>
      <Box
        component={RouterLink}
        sx={{
          color: "inherit",
          display: "block",
          height: "100%",
          textDecoration: "none",
          "&:focus-visible .CategoryLandingItem-surface": {
            outline: "2px solid",
            outlineColor: "primary.main",
            outlineOffset: 2,
          },
          "&:hover .CategoryLandingItem-surface": {
            borderColor: "primary.main",
            boxShadow: 4,
            transform: "translateY(-2px)",
          },
        }}
        to={card.path}
      >
        <Paper
          className="CategoryLandingItem-surface"
          sx={{
            height: "100%",
            p: 3,
            transition: (theme) => {
              return theme.transitions.create(["border-color", "box-shadow", "transform"], {
                duration: theme.transitions.duration.short,
              });
            },
          }}
        >
          <Stack height="100%" spacing={1.5}>
            <Stack alignItems="center" direction="row" spacing={1.25}>
              <Box color="primary.main" flexShrink={0} lineHeight={0}>
                {card.icon}
              </Box>
              <Typography variant="h2">{card.title}</Typography>
            </Stack>
            <Typography color="text.secondary" variant="body2">
              {card.description}
            </Typography>
          </Stack>
        </Paper>
      </Box>
    </Grid>
  );
};
