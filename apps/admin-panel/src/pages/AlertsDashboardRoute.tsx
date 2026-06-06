import type { AlertCategory, AlertDomain } from "@dto/alerts/alerts";
import { AlertsDashboardPage } from "@pages/AlertsDashboardPage";
import { AlertsControlPanelProvider } from "@providers/AlertsControlPanelProvider";
import { AlertsProvider } from "@providers/AlertsProvider";
import type { ReactElement } from "react";

/**
 * Static props required to bind one alerts dashboard route slice.
 */
type AlertsDashboardRouteProps = {
  /**
   * Route category resolved statically by the alerts route module.
   */
  category: AlertCategory;
  /**
   * Route domain resolved statically by the alerts route module.
   */
  domain: AlertDomain;
};

/**
 * Wires one concrete alerts route slice to shared data and control-panel providers.
 */
export const AlertsDashboardRoute = ({
  category,
  domain,
}: AlertsDashboardRouteProps): ReactElement => {
  return (
    <AlertsProvider category={category} domain={domain}>
      <AlertsControlPanelProvider>
        <AlertsDashboardPage />
      </AlertsControlPanelProvider>
    </AlertsProvider>
  );
};
