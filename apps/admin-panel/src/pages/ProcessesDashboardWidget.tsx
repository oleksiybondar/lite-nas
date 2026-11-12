import { formatMetricBytes } from "@helpers/metric-display";
import { useProcessMetric } from "@hooks/useProcessMetric";
import { useLanguage } from "@lite-nas/language";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Shows the highest CPU processes from the current process snapshot. */
export const ProcessesDashboardWidget = (): ReactElement => {
  const { snapshot } = useProcessMetric();
  const { getLabel } = useLanguage();
  const processes = [...(snapshot?.processes ?? [])]
    .sort((left, right) => right.cpu.cpu_pct - left.cpu.cpu_pct)
    .slice(0, 5);
  return (
    <Stack spacing={1}>
      {processes.map((process) => (
        <Stack direction="row" justifyContent="space-between" key={process.pid} spacing={1}>
          <Typography>
            {process.name} ({process.pid})
          </Typography>
          <Typography>
            {process.cpu.cpu_pct.toFixed(1)}% · {formatMetricBytes(process.memory.rss_bytes)}
          </Typography>
        </Stack>
      ))}
      {processes.length === 0 && (
        <Typography color="text.secondary">{getLabel("admin.dashboard.noData")}</Typography>
      )}
    </Stack>
  );
};
