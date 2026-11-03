import { useServiceMetric } from "@hooks/useServiceMetric";
import { useLanguage } from "@lite-nas/language";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Summarizes active services and lists services needing attention. */
export const ServicesDashboardWidget = (): ReactElement => {
  const { snapshot } = useServiceMetric();
  const { getLabel } = useLanguage();
  const services = snapshot?.services ?? [];
  const active = services.filter((service) => service.active_state === "active").length;
  const inactive = services.filter((service) => service.active_state !== "active").slice(0, 5);
  return (
    <Stack spacing={1}>
      <Typography>
        {getLabel("admin.dashboard.activeServices")}: {active}/{services.length}
      </Typography>
      {inactive.map((service) => (
        <Stack
          direction="row"
          justifyContent="space-between"
          key={`${service.manager}:${service.name}`}
          spacing={1}
        >
          <Typography>{service.name}</Typography>
          <Typography color="text.secondary">{service.active_state ?? "—"}</Typography>
        </Stack>
      ))}
    </Stack>
  );
};
