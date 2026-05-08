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
    <Grid
      data-test-class="category-landing-card-grid-item"
      data-test-name={card.title}
      data-test-path={card.path}
      size={{ md: 4, sm: 6, xs: 12 }}
    >
      <Box
        component={RouterLink}
        data-test-class="category-landing-card-link"
        data-test-name={card.title}
        data-test-path={card.path}
        sx={categoryLandingItemLinkSx}
        to={card.path}
      >
        {renderCategoryLandingCard(card)}
      </Box>
    </Grid>
  );
};

/**
 * Link-state styles for the clickable category card wrapper.
 */
const categoryLandingItemLinkSx = {
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
};

/**
 * Builds the visible category card surface.
 */
const renderCategoryLandingCard = (card: CategoryLandingCard): ReactElement => {
  return (
    <Paper
      className="CategoryLandingItem-surface"
      data-test-class="category-landing-card"
      data-test-name={card.title}
      data-test-path={card.path}
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
        {renderCategoryLandingCardHeader(card)}
        <Typography
          color="text.secondary"
          data-test-class="category-landing-card-description"
          data-test-name={card.title}
          variant="body2"
        >
          {card.description}
        </Typography>
      </Stack>
    </Paper>
  );
};

/**
 * Builds the icon and title row for a category card.
 */
const renderCategoryLandingCardHeader = (card: CategoryLandingCard): ReactElement => {
  return (
    <Stack alignItems="center" direction="row" spacing={1.25}>
      <Box
        color="primary.main"
        data-test-class="category-landing-card-icon"
        data-test-name={card.title}
        flexShrink={0}
        lineHeight={0}
      >
        {card.icon}
      </Box>
      <Typography
        data-test-class="category-landing-card-title"
        data-test-name={card.title}
        variant="h2"
      >
        {card.title}
      </Typography>
    </Stack>
  );
};
