import { ChartSvgContainer } from "@components/monitoring/ChartSvgContainer";
import { MonitoringChartTooltip } from "@components/monitoring/MonitoringChartTooltip";
import { formatChartStamp } from "@components/monitoring/percent-chart-shared";
import type { ReactElement } from "react";
import {
  formatPercentGradientMultiChartValue,
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
   * Maximum number of values represented by the fixed X scale.
   */
  capacity: number;
  /**
   * Precomputed line paths and colors for each rendered series.
   */
  lines: PercentGradientMultiChartLine[];
  /**
   * Ordered timestamps associated with each plotted point.
   */
  stamps: string[];
  /**
   * Ordered percent series plotted on a fixed 0-100 vertical scale.
   */
  valuesByKey: Record<string, number[]>;
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
  capacity,
  lines,
  stamps,
  valuesByKey,
}: PercentGradientMultiChartSvgProps): ReactElement => {
  return (
    <ChartSvgContainer
      ariaLabel="Percent gradient multi chart"
      capacity={capacity}
      frame={percentGradientMultiChartFrame}
      length={stamps.length}
      testId="percent-gradient-multi-chart"
      tooltip={(hoveredPoint) => {
        if (hoveredPoint === null) {
          return null;
        }

        return (
          <MonitoringChartTooltip
            chartWidth={percentGradientMultiChartFrame.chartWidth}
            items={lines.map((line) => {
              return {
                color: line.color,
                label: line.key,
                value: formatPercentGradientMultiChartValue(
                  valuesByKey[line.key]?.[hoveredPoint.index] ?? 0,
                ),
              };
            })}
            label={formatChartStamp(stamps[hoveredPoint.index] ?? "")}
            x={hoveredPoint.x}
          />
        );
      }}
    >
      {renderPercentGradientMultiChartGrid()}
      {renderPercentGradientMultiChartAxisLabels(axisLabels)}
      {renderPercentGradientMultiChartLines(lines)}
    </ChartSvgContainer>
  );
};
