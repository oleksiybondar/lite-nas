import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import {
  buildPercentGradientChartAreaPath,
  buildPercentGradientChartAxisLabels,
  buildPercentGradientChartLatestGuideY,
  buildPercentGradientChartLinePath,
  hasPercentGradientChartData,
} from "./helpers";
import { PercentGradientChartSvg } from "./PercentGradientChartSvg";
import type { PercentGradientChartProps } from "./types";

/**
 * Fixed-scale percent chart with quarter-band gradient coloring.
 * Titles and surrounding card layout stay with the parent consumer.
 */
export const PercentGradientChart = ({
  capacity,
  heightPx,
  stamps,
  values,
}: PercentGradientChartProps): ReactElement => {
  if (!hasPercentGradientChartData(values, stamps)) {
    return (
      <Typography color="text.secondary" data-testid="percent-gradient-chart-empty" variant="body2">
        Chart data will appear after telemetry points are loaded.
      </Typography>
    );
  }

  return (
    <PercentGradientChartSvg
      areaPath={buildPercentGradientChartAreaPath(values, capacity)}
      axisLabels={buildPercentGradientChartAxisLabels(stamps, capacity)}
      capacity={capacity}
      heightPx={heightPx}
      latestGuideY={buildPercentGradientChartLatestGuideY(values)}
      linePath={buildPercentGradientChartLinePath(values, capacity)}
      stamps={stamps}
      values={values}
    />
  );
};
