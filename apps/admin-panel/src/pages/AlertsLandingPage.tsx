import {
  type CategoryLandingCard,
  CategoryLandingPage,
} from "@components/navigation/CategoryLandingPage";
import { useRbac } from "@hooks/useRbac";
import MemoryIcon from "@mui/icons-material/Memory";
import SecurityIcon from "@mui/icons-material/Security";
import type { ReactElement } from "react";

/**
 * Landing page for alert domain overviews.
 */
export const AlertsLandingPage = (): ReactElement => {
  const { requireOperator, requireSecurity } = useRbac();

  return (
    <CategoryLandingPage
      cards={buildAlertDomainCards(requireOperator(), requireSecurity())}
      overline="Alerts"
      summary="Browse operator and security alert dashboards by domain before drilling into status-specific queues."
      title="Alerts"
    />
  );
};

/**
 * Builds the alert-domain cards available to the current principal.
 */
const buildAlertDomainCards = (
  showSystemAlerts: boolean,
  showSecurityAlerts: boolean,
): CategoryLandingCard[] => {
  const cards: CategoryLandingCard[] = [];

  if (showSystemAlerts) {
    cards.push({
      description: "Operational alerts for infrastructure, services, and runtime health.",
      icon: <MemoryIcon />,
      path: "/alerts/system",
      title: "System",
    });
  }

  if (showSecurityAlerts) {
    cards.push({
      description: "Security alerts for access, policy, and protection events.",
      icon: <SecurityIcon />,
      path: "/alerts/security",
      title: "Security",
    });
  }

  return cards;
};
