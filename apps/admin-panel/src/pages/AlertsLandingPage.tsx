import {
  type CategoryLandingCard,
  CategoryLandingPage,
} from "@components/navigation/CategoryLandingPage";
import { useRbac } from "@hooks/useRbac";
import { useLanguage } from "@lite-nas/language";
import MemoryIcon from "@mui/icons-material/Memory";
import SecurityIcon from "@mui/icons-material/Security";
import type { ReactElement } from "react";

/**
 * Landing page for alert domain overviews.
 */
export const AlertsLandingPage = (): ReactElement => {
  const { requireOperator, requireSecurity } = useRbac();
  const { getLabel } = useLanguage();

  return (
    <CategoryLandingPage
      cards={buildAlertDomainCards(requireOperator(), requireSecurity(), getLabel)}
      overline={getLabel("navigation.Alerts")}
      summary={getLabel("admin.landing.alertsSummary")}
      title={getLabel("navigation.Alerts")}
    />
  );
};

/**
 * Builds the alert-domain cards available to the current principal.
 */
const buildAlertDomainCards = (
  showSystemAlerts: boolean,
  showSecurityAlerts: boolean,
  getLabel: (key: string) => string,
): CategoryLandingCard[] => {
  const cards: CategoryLandingCard[] = [];

  if (showSystemAlerts) {
    cards.push({
      description: getLabel("admin.landing.alertsSystem"),
      icon: <MemoryIcon />,
      path: "/alerts/system",
      title: getLabel("navigation.System"),
    });
  }

  if (showSecurityAlerts) {
    cards.push({
      description: getLabel("admin.landing.alertsSecurity"),
      icon: <SecurityIcon />,
      path: "/alerts/security",
      title: getLabel("navigation.Security"),
    });
  }

  return cards;
};
