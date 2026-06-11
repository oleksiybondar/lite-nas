import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import type { PercentGradientMultiChartLegendItem } from "./types";

type PercentGradientMultiChartLegendProps = {
  /**
   * Legend entries aligned with the rendered chart lines.
   */
  items: PercentGradientMultiChartLegendItem[];
};

/**
 * Compact legend showing the latest value for each rendered series.
 */
export const PercentGradientMultiChartLegend = ({
  items,
}: PercentGradientMultiChartLegendProps): ReactElement => {
  return (
    <Stack
      data-test-class="percent-gradient-multi-chart-legend"
      direction="row"
      flexWrap="wrap"
      gap={1.5}
      useFlexGap
    >
      {items.map((item) => {
        return (
          <Typography
            color={item.color}
            data-test-class="percent-gradient-multi-chart-legend-item"
            data-test-name={item.key}
            key={item.key}
            variant="body2"
          >
            {item.key}: {Math.round(item.latestValue)}%
          </Typography>
        );
      })}
    </Stack>
  );
};
