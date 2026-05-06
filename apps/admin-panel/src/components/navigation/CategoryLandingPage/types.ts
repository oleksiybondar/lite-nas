import type { ReactNode } from "react";

/**
 * Card shown on category landing pages.
 */
export type CategoryLandingCard = {
  /**
   * Summary that explains what the target area contains.
   */
  description: string;
  /**
   * Icon shown before the card title.
   */
  icon: ReactNode;
  /**
   * Route path opened by the card.
   */
  path: string;
  /**
   * Card title.
   */
  title: string;
};
