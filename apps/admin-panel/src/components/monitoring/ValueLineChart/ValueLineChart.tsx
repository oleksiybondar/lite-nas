import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import {
  buildValueLineChartAxisLabels,
  buildValueLineChartLegendItems,
  buildValueLineChartLines,
  buildValueLineChartYAxisLabels,
  formatValueLineChartValue,
  hasValueLineChartData,
} from "./helpers";
import type { ValueLineChartProps } from "./types";
import { ValueLineChartLegend } from "./ValueLineChartLegend";
import { ValueLineChartSvg } from "./ValueLineChartSvg";

/**
 * Dynamic multi-series value chart with one colored line per series.
 * Titles and surrounding card layout stay with the parent consumer.
 */
export const ValueLineChart = ({
  capacity,
  colorStops,
  formatValue = formatValueLineChartValue,
  guideStops,
  stamps,
  valuesByKey,
}: ValueLineChartProps): ReactElement => {
  const minimumMaximum = Math.max(
    0,
    ...(colorStops ?? []).map((stop) => stop.value),
    ...(guideStops ?? [])
      .filter((stop) => stop.showBar || stop.showLabel)
      .map((stop) => stop.value),
  );
  if (!hasValueLineChartData(valuesByKey, stamps)) {
    return (
      <Typography color="text.secondary" data-testid="value-line-chart-empty" variant="body2">
        Chart data will appear after telemetry points are loaded.
      </Typography>
    );
  }

  return (
    <>
      <ValueLineChartLegend
        formatValue={formatValue}
        items={buildValueLineChartLegendItems(valuesByKey)}
      />
      <ValueLineChartSvg
        axisLabels={buildValueLineChartAxisLabels(stamps, capacity)}
        capacity={capacity}
        colorStops={colorStops}
        formatValue={formatValue}
        guideStops={guideStops}
        lines={buildValueLineChartLines(valuesByKey, capacity, minimumMaximum)}
        stamps={stamps}
        valuesByKey={valuesByKey}
        yAxisLabels={buildValueLineChartYAxisLabels(valuesByKey, formatValue, minimumMaximum)}
      />
    </>
  );
};
