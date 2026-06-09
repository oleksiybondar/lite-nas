import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import type { ValueLineChartLegendItem } from "./types";

type ValueLineChartLegendProps = {
  /**
   * Optional formatter used by legend values.
   */
  formatValue: (value: number) => string;
  /**
   * Legend entries aligned with the rendered chart lines.
   */
  items: ValueLineChartLegendItem[];
};

/**
 * Compact legend showing the latest value for each rendered series.
 */
export const ValueLineChartLegend = ({
  formatValue,
  items,
}: ValueLineChartLegendProps): ReactElement => {
  return (
    <Stack
      data-test-class="value-line-chart-legend"
      direction="row"
      flexWrap="wrap"
      gap={1.5}
      useFlexGap
    >
      {items.map((item) => {
        return (
          <Typography
            color={item.color}
            data-test-class="value-line-chart-legend-item"
            data-test-name={item.key}
            key={item.key}
            variant="body2"
          >
            {item.key}: {formatValue(item.latestValue)}
          </Typography>
        );
      })}
    </Stack>
  );
};
