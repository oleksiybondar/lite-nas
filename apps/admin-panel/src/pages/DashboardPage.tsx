import DnsRoundedIcon from "@mui/icons-material/DnsRounded";
import InsightsRoundedIcon from "@mui/icons-material/InsightsRounded";
import SecurityRoundedIcon from "@mui/icons-material/SecurityRounded";
import Box from "@mui/material/Box";
import Grid from "@mui/material/Grid";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

const dashboardSections = [
  {
    description: "Runtime health and host telemetry will land here.",
    icon: <InsightsRoundedIcon color="primary" />,
    title: "System Metrics",
  },
  {
    description: "Operator and permission workflows can plug into this area.",
    icon: <SecurityRoundedIcon color="primary" />,
    title: "Access",
  },
  {
    description: "Storage, services, and gateway state can be added as slices.",
    icon: <DnsRoundedIcon color="primary" />,
    title: "Services",
  },
];

export const DashboardPage = (): ReactElement => {
  return (
    <Stack data-testid="dashboard-page" spacing={4}>
      <Stack data-testid="dashboard-page-header" maxWidth="760px" spacing={1}>
        <Typography color="primary" data-testid="dashboard-page-overline" variant="overline">
          Admin panel
        </Typography>
        <Typography data-testid="dashboard-page-title" variant="h1">
          LiteNAS operations
        </Typography>
        <Typography color="text.secondary" data-testid="dashboard-page-summary" variant="body1">
          Initial browser shell for the LiteNAS administration experience.
        </Typography>
      </Stack>

      <Grid container data-testid="dashboard-section-list" spacing={2}>
        {dashboardSections.map((section) => {
          return (
            <Grid
              data-test-class="dashboard-section-grid-item"
              data-test-name={section.title}
              key={section.title}
              size={{ md: 4, xs: 12 }}
            >
              <Paper
                data-test-class="dashboard-section-card"
                data-test-name={section.title}
                sx={{ height: "100%", p: 3 }}
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
            </Grid>
          );
        })}
      </Grid>
    </Stack>
  );
};
