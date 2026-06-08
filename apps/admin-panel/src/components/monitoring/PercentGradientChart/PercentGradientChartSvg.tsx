import Box from "@mui/material/Box";
import type { ReactElement } from "react";
import {
  mapPercentGradientChartGuideY,
  type PercentGradientChartAxisLabel,
  percentGradientChartFrame,
} from "./helpers";

type PercentGradientChartSvgProps = {
  /**
   * Closed area path rendered under the percent chart line.
   */
  areaPath: string;
  /**
   * X-axis labels positioned against the full chart capacity scale.
   */
  axisLabels: PercentGradientChartAxisLabel[];
  /**
   * Optional rendered chart height in pixels.
   */
  heightPx?: number;
  /**
   * Y coordinate of the latest-value guide rendered across the chart body.
   */
  latestGuideY: number | null;
  /**
   * Open line path rendered above the gradient fill.
   */
  linePath: string;
};

const gradientStops = [
  { color: "#2563eb", offset: "0%" },
  { color: "#2563eb", offset: "25%" },
  { color: "#16a34a", offset: "50%" },
  { color: "#ea580c", offset: "75%" },
  { color: "#dc2626", offset: "100%" },
] as const;
const gridValues = [0, 25, 50, 75, 100] as const;

/**
 * Renders the fixed horizontal guides and Y-axis labels for the percent chart.
 */
const renderPercentGradientChartGrid = (): ReactElement[] => {
  return gridValues.map((gridValue) => {
    return (
      <g key={gridValue}>
        <line
          stroke="rgba(148, 163, 184, 0.35)"
          strokeDasharray={gridValue === 0 ? undefined : "4 6"}
          strokeWidth="1"
          x1={percentGradientChartFrame.leftPadding}
          x2={percentGradientChartFrame.chartWidth - percentGradientChartFrame.rightPadding}
          y1={mapPercentGradientChartGuideY(gridValue)}
          y2={mapPercentGradientChartGuideY(gridValue)}
        />
        <text
          fill="currentColor"
          fontSize="11"
          textAnchor="end"
          x={percentGradientChartFrame.leftPadding - 8}
          y={mapPercentGradientChartGuideY(gridValue) + 4}
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
const renderPercentGradientChartAxisLabels = (
  axisLabels: PercentGradientChartAxisLabel[],
): ReactElement[] => {
  return axisLabels.map((label) => {
    return (
      <text
        fill="currentColor"
        fontSize="11"
        key={`${label.x}-${label.label}`}
        textAnchor="middle"
        x={label.x}
        y={percentGradientChartFrame.chartHeight - 8}
      >
        {label.label}
      </text>
    );
  });
};

/**
 * SVG frame rendered by the percent gradient chart component.
 */
export const PercentGradientChartSvg = ({
  areaPath,
  axisLabels,
  heightPx,
  latestGuideY,
  linePath,
}: PercentGradientChartSvgProps): ReactElement => {
  return (
    <Box
      data-testid="percent-gradient-chart"
      sx={{ height: heightPx, overflow: "visible", width: "100%" }}
    >
      <svg
        aria-label="Percent gradient chart"
        height={heightPx}
        preserveAspectRatio="none"
        viewBox={`0 0 ${percentGradientChartFrame.chartWidth} ${percentGradientChartFrame.chartHeight}`}
        width="100%"
      >
        <defs>
          <linearGradient id="percent-gradient-chart-fill" x1="0" x2="0" y1="1" y2="0">
            {gradientStops.map((stop) => {
              return <stop key={stop.offset} offset={stop.offset} stopColor={stop.color} />;
            })}
          </linearGradient>
        </defs>
        {renderPercentGradientChartGrid()}
        {renderPercentGradientChartAxisLabels(axisLabels)}
        {latestGuideY !== null ? (
          <line
            stroke="rgba(15, 23, 42, 0.45)"
            strokeDasharray="6 6"
            strokeWidth="1.5"
            x1={percentGradientChartFrame.leftPadding}
            x2={percentGradientChartFrame.chartWidth - percentGradientChartFrame.rightPadding}
            y1={latestGuideY}
            y2={latestGuideY}
          />
        ) : null}
        <path d={areaPath} fill="url(#percent-gradient-chart-fill)" fillOpacity="0.26" />
        <path d={linePath} fill="none" stroke="rgba(15, 23, 42, 0.88)" strokeWidth="2.5" />
      </svg>
    </Box>
  );
};
