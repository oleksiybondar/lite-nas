import { useLanguage } from "@lite-nas/language";
import AppsRoundedIcon from "@mui/icons-material/AppsRounded";
import DnsRoundedIcon from "@mui/icons-material/DnsRounded";
import InsightsRoundedIcon from "@mui/icons-material/InsightsRounded";
import SettingsRoundedIcon from "@mui/icons-material/SettingsRounded";
import Box from "@mui/material/Box";
import Grid from "@mui/material/Grid";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { Link as RouterLink } from "react-router-dom";

type DashboardSection = {
  description: string;
  icon: ReactElement;
  path: string;
  title: string;
};

/** Presents direct navigation to the main admin panel areas. */
export const DashboardPage = (): ReactElement => {
  const { getLabel } = useLanguage();
  const dashboardSections: DashboardSection[] = [
    {
      description: getLabel("admin.landing.dashboardSystemDesc"),
      icon: <InsightsRoundedIcon color="primary" />,
      path: "/system/performance",
      title: getLabel("admin.landing.dashboardSystemTitle"),
    },
    {
      description: getLabel("admin.landing.dashboardServicesDesc"),
      icon: <DnsRoundedIcon color="primary" />,
      path: "/system/processes/services",
      title: getLabel("navigation.Services"),
    },
    {
      description: getLabel("admin.landing.dashboardApplicationsDesc"),
      icon: <AppsRoundedIcon color="primary" />,
      path: "/applications",
      title: getLabel("navigation.Applications"),
    },
    {
      description: getLabel("admin.landing.dashboardPreferencesDesc"),
      icon: <SettingsRoundedIcon color="primary" />,
      path: "/preferences",
      title: getLabel("navigation.Preferences"),
    },
  ];
  return (
    <Stack data-testid="dashboard-page" spacing={4}>
      <Stack data-testid="dashboard-page-header" maxWidth="760px" spacing={1}>
        <Typography color="primary" data-testid="dashboard-page-overline" variant="overline">
          {getLabel("admin.landing.dashboardOverline")}
        </Typography>
        <Typography data-testid="dashboard-page-title" variant="h1">
          {getLabel("admin.landing.dashboardTitle")}
        </Typography>
        <Typography color="text.secondary" data-testid="dashboard-page-summary" variant="body1">
          {getLabel("admin.landing.dashboardSummary")}
        </Typography>
      </Stack>

      <Grid container data-testid="dashboard-section-list" spacing={2}>
        {dashboardSections.map(renderDashboardSection)}
      </Grid>
    </Stack>
  );
};

/** Renders one whole-card link to an admin area. */
const renderDashboardSection = (section: DashboardSection): ReactElement => (
  <Grid
    data-test-class="dashboard-section-grid-item"
    data-test-name={section.title}
    data-test-path={section.path}
    key={section.title}
    size={{ md: 4, xs: 12 }}
  >
    <Box
      component={RouterLink}
      data-test-class="dashboard-section-link"
      data-test-name={section.title}
      data-test-path={section.path}
      sx={dashboardSectionLinkSx}
      to={section.path}
    >
      <Paper
        className="DashboardSection-surface"
        data-test-class="dashboard-section-card"
        data-test-name={section.title}
        data-test-path={section.path}
        sx={{
          height: "100%",
          p: 3,
          transition: (theme) =>
            theme.transitions.create(["border-color", "box-shadow", "transform"], {
              duration: theme.transitions.duration.short,
            }),
        }}
      >
        <Stack spacing={2}>
          <Box data-test-class="dashboard-section-icon" data-test-name={section.title}>
            {section.icon}
          </Box>
          <Stack spacing={0.75}>
            <Typography
              data-test-class="dashboard-section-title"
              data-test-name={section.title}
              variant="h2"
            >
              {section.title}
            </Typography>
            <Typography
              color="text.secondary"
              data-test-class="dashboard-section-description"
              data-test-name={section.title}
              variant="body2"
            >
              {section.description}
            </Typography>
          </Stack>
        </Stack>
      </Paper>
    </Box>
  </Grid>
);

/** Gives dashboard links the same whole-card interaction as category pages. */
const dashboardSectionLinkSx = {
  color: "inherit",
  display: "block",
  height: "100%",
  textDecoration: "none",
  "&:focus-visible .DashboardSection-surface": {
    outline: "2px solid",
    outlineColor: "primary.main",
    outlineOffset: 2,
  },
  "&:hover .DashboardSection-surface": {
    borderColor: "primary.main",
    boxShadow: 4,
    transform: "translateY(-2px)",
  },
};
