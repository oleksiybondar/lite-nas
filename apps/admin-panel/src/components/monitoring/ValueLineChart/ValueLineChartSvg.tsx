import { ChartSvgContainer } from "@components/monitoring/ChartSvgContainer";
import { MonitoringChartTooltip } from "@components/monitoring/MonitoringChartTooltip";
import { formatChartStamp } from "@components/monitoring/percent-chart-shared";
import type { ReactElement } from "react";
import { valueLineChartFrame } from "./helpers";
import type {
  ValueLineChartAxisLabel,
  ValueLineChartLine,
  ValueLineChartYAxisLabel,
} from "./types";

type ValueLineChartSvgProps = {
  /**
   * X-axis labels positioned against the full chart capacity scale.
   */
  axisLabels: ValueLineChartAxisLabel[];
  /**
   * Maximum number of values represented by the fixed X scale.
   */
  capacity: number;
  /**
   * Formatter reused for hover tooltip values.
   */
  formatValue: (value: number) => string;
  /**
   * Precomputed line paths and colors for each rendered series.
   */
  lines: ValueLineChartLine[];
  /**
   * Ordered timestamps associated with each plotted point.
   */
  stamps: string[];
  /**
   * Ordered value series plotted on a dynamic vertical scale.
   */
  valuesByKey: Record<string, number[]>;
  /**
   * Y-axis labels and guide positions for the current dynamic domain.
   */
  yAxisLabels: ValueLineChartYAxisLabel[];
};

/**
 * Renders the dynamic horizontal guides and Y-axis labels for the value chart.
 */
const renderValueLineChartGrid = (yAxisLabels: ValueLineChartYAxisLabel[]): ReactElement[] => {
  return yAxisLabels.map((label, index) => {
    return (
      <g key={`${label.value}-${label.label}`}>
        <line
          stroke="rgba(148, 163, 184, 0.35)"
          strokeDasharray={index === yAxisLabels.length - 1 ? undefined : "4 6"}
          strokeWidth="1"
          x1={valueLineChartFrame.leftPadding}
          x2={valueLineChartFrame.chartWidth - valueLineChartFrame.rightPadding}
          y1={label.y}
          y2={label.y}
        />
        <text
          fill="currentColor"
          fontSize="12"
          textAnchor="end"
          x={valueLineChartFrame.leftPadding - 8}
          y={label.y + 4}
        >
          {label.label}
        </text>
      </g>
    );
  });
};

/**
 * Renders X-axis labels that align with the fixed-capacity timestamp scale.
 */
const renderValueLineChartAxisLabels = (axisLabels: ValueLineChartAxisLabel[]): ReactElement[] => {
  return axisLabels.map((label) => {
    return (
      <text
        fill="currentColor"
        fontSize="12"
        key={`${label.x}-${label.label}`}
        textAnchor="middle"
        x={label.x}
        y={valueLineChartFrame.chartHeight - 8}
      >
        {label.label}
      </text>
    );
  });
};

/**
 * Renders one SVG path per colored value series.
 */
const renderValueLineChartLines = (lines: ValueLineChartLine[]): ReactElement[] => {
  return lines.map((line) => {
    return <path d={line.path} fill="none" key={line.key} stroke={line.color} strokeWidth="2" />;
  });
};

/**
 * SVG frame rendered by the dynamic value line chart component.
 */
export const ValueLineChartSvg = ({
  axisLabels,
  capacity,
  formatValue,
  lines,
  stamps,
  valuesByKey,
  yAxisLabels,
}: ValueLineChartSvgProps): ReactElement => {
  return (
    <ChartSvgContainer
      ariaLabel="Value line chart"
      capacity={capacity}
      frame={valueLineChartFrame}
      length={stamps.length}
      testId="value-line-chart"
      tooltip={(hoveredPoint) => {
        if (hoveredPoint === null) {
          return null;
        }

        const hoveredStamp = stamps[hoveredPoint.index];

        if (hoveredStamp === undefined) {
          return null;
        }

        return (
          <MonitoringChartTooltip
            chartWidth={valueLineChartFrame.chartWidth}
            items={lines.flatMap((line) => {
              const value = valuesByKey[line.key]?.[hoveredPoint.index];

              if (value === undefined) {
                return [];
              }

              return [
                {
                  color: line.color,
                  label: line.key,
                  value: formatValue(value),
                },
              ];
            })}
            label={formatChartStamp(hoveredStamp)}
            x={hoveredPoint.x}
          />
        );
      }}
    >
      {renderValueLineChartGrid(yAxisLabels)}
      {renderValueLineChartAxisLabels(axisLabels)}
      {renderValueLineChartLines(lines)}
    </ChartSvgContainer>
  );
};
