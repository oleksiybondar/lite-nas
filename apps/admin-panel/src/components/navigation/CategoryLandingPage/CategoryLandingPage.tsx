import Grid from "@mui/material/Grid";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { CategoryLandingItem } from "./CategoryLandingItem";
import type { CategoryLandingCard } from "./types";

type CategoryLandingPageProps = {
  /**
   * Cards that expose child routes or subcategories.
   */
  cards: CategoryLandingCard[];
  /**
   * Main page heading.
   */
  title: string;
  /**
   * Small label rendered above the heading.
   */
  overline: string;
  /**
   * Introductory description for the category.
   */
  summary: string;
};

/**
 * Reusable overview page for route categories.
 */
export const CategoryLandingPage = ({
  cards,
  overline,
  summary,
  title,
}: CategoryLandingPageProps): ReactElement => {
  return (
    <Stack data-testid="category-landing-page" data-test-name={title} spacing={4}>
      <Stack data-testid="category-landing-header" maxWidth="820px" spacing={1}>
        <Typography color="primary" data-testid="category-landing-overline" variant="overline">
          {overline}
        </Typography>
        <Typography data-testid="category-landing-title" variant="h1">
          {title}
        </Typography>
        <Typography color="text.secondary" data-testid="category-landing-summary" variant="body1">
          {summary}
        </Typography>
      </Stack>

      <Grid container data-testid="category-landing-card-list" spacing={2}>
        {cards.map((card) => {
          return <CategoryLandingItem card={card} key={card.path} />;
        })}
      </Grid>
    </Stack>
  );
};
