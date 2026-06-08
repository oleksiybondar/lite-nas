import Box from "@mui/material/Box";
import type { ReactElement } from "react";
import {
  mapPercentGradientMultiChartGuideY,
  percentGradientMultiChartFrame,
  percentGradientMultiChartGridValues,
} from "./helpers";
import type { PercentGradientMultiChartAxisLabel, PercentGradientMultiChartLine } from "./types";

type PercentGradientMultiChartSvgProps = {
  /**
   * X-axis labels positioned against the full chart capacity scale.
   */
  axisLabels: PercentGradientMultiChartAxisLabel[];
  /**
   * Precomputed line paths and colors for each rendered series.
   */
  lines: PercentGradientMultiChartLine[];
};

/**
 * Renders the fixed horizontal guides and Y-axis labels for the multi-series chart.
 */
const renderPercentGradientMultiChartGrid = (): ReactElement[] => {
  return percentGradientMultiChartGridValues.map((gridValue) => {
    return (
      <g key={gridValue}>
        <line
          stroke="rgba(148, 163, 184, 0.35)"
          strokeDasharray={gridValue === 0 ? undefined : "4 6"}
          strokeWidth="1"
          x1={percentGradientMultiChartFrame.leftPadding}
          x2={
            percentGradientMultiChartFrame.chartWidth - percentGradientMultiChartFrame.rightPadding
          }
          y1={mapPercentGradientMultiChartGuideY(gridValue)}
          y2={mapPercentGradientMultiChartGuideY(gridValue)}
        />
        <text
          fill="currentColor"
          fontSize="11"
          textAnchor="end"
          x={percentGradientMultiChartFrame.leftPadding - 8}
          y={mapPercentGradientMultiChartGuideY(gridValue) + 4}
        >
          {gridValue}%
        </text>
      </g>
    );
  });
};

/**
 * Renders X-axis labels that align with the fixed-capacity timestamp scale.
 */
const renderPercentGradientMultiChartAxisLabels = (
  axisLabels: PercentGradientMultiChartAxisLabel[],
): ReactElement[] => {
  return axisLabels.map((label) => {
    return (
      <text
        fill="currentColor"
        fontSize="11"
        key={`${label.x}-${label.label}`}
        textAnchor="middle"
        x={label.x}
        y={percentGradientMultiChartFrame.chartHeight - 8}
      >
        {label.label}
      </text>
    );
  });
};

/**
 * Renders one SVG path per colored percent series.
 */
const renderPercentGradientMultiChartLines = (
  lines: PercentGradientMultiChartLine[],
): ReactElement[] => {
  return lines.map((line) => {
    return <path d={line.path} fill="none" key={line.key} stroke={line.color} strokeWidth="2" />;
  });
};

/**
 * SVG frame rendered by the multi-series percent gradient chart component.
 */
export const PercentGradientMultiChartSvg = ({
  axisLabels,
  lines,
}: PercentGradientMultiChartSvgProps): ReactElement => {
  return (
    <Box data-testid="percent-gradient-multi-chart" sx={{ overflow: "visible", width: "100%" }}>
      <svg
        aria-label="Percent gradient multi chart"
        preserveAspectRatio="none"
        viewBox={`0 0 ${percentGradientMultiChartFrame.chartWidth} ${percentGradientMultiChartFrame.chartHeight}`}
        width="100%"
      >
        {renderPercentGradientMultiChartGrid()}
        {renderPercentGradientMultiChartAxisLabels(axisLabels)}
        {renderPercentGradientMultiChartLines(lines)}
      </svg>
    </Box>
  );
};
