import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import {
  buildPercentGradientMultiChartAxisLabels,
  buildPercentGradientMultiChartLegendItems,
  buildPercentGradientMultiChartLines,
  hasPercentGradientMultiChartData,
} from "./helpers";
import { PercentGradientMultiChartLegend } from "./PercentGradientMultiChartLegend";
import { PercentGradientMultiChartSvg } from "./PercentGradientMultiChartSvg";
import type { PercentGradientMultiChartProps } from "./types";

/**
 * Fixed-scale multi-series percent chart with one colored line per series.
 * Titles and surrounding card layout stay with the parent consumer.
 */
export const PercentGradientMultiChart = ({
  capacity,
  stamps,
  valuesByKey,
}: PercentGradientMultiChartProps): ReactElement => {
  if (!hasPercentGradientMultiChartData(valuesByKey, stamps)) {
    return (
      <Typography
        color="text.secondary"
        data-testid="percent-gradient-multi-chart-empty"
        variant="body2"
      >
        Chart data will appear after telemetry points are loaded.
      </Typography>
    );
  }

  return (
    <>
      <PercentGradientMultiChartLegend
        items={buildPercentGradientMultiChartLegendItems(valuesByKey)}
      />
      <PercentGradientMultiChartSvg
        axisLabels={buildPercentGradientMultiChartAxisLabels(stamps, capacity)}
        lines={buildPercentGradientMultiChartLines(valuesByKey, capacity)}
      />
    </>
  );
};
