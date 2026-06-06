import type { AlertCategory, AlertDomain } from "@dto/alerts/alerts";
import { AlertsProvider } from "@providers/AlertsProvider";
import type { ReactElement } from "react";
import { AlertsDashboardPage } from "./AlertsDashboardPage";

type AlertsDashboardRouteProps = {
  /**
   * Current alerts route category.
   */
  category: AlertCategory;
  /**
   * Current alerts route domain.
   */
  domain: AlertDomain;
};

/**
 * Route adapter that binds one dashboard page to a concrete alerts provider slice.
 */
export const AlertsDashboardRoute = ({
  category,
  domain,
}: AlertsDashboardRouteProps): ReactElement => {
  return (
    <AlertsProvider category={category} domain={domain}>
      <AlertsDashboardPage />
    </AlertsProvider>
  );
};
