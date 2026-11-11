import { ChartSvgContainer } from "@components/monitoring/ChartSvgContainer";
import { MonitoringChartTooltip } from "@components/monitoring/MonitoringChartTooltip";
import {
  type ChartHoverPoint,
  formatChartStamp,
} from "@components/monitoring/percent-chart-shared";
import { type ReactElement, useId } from "react";
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
  /** Optional value/color stops for a single plotted series. */
  colorStops?: { value: number; color: string }[] | undefined;
  /**
   * Formatter reused for hover tooltip values.
   */
  formatValue: (value: number) => string;
  guideStops?:
    | { value: number; color: string; label: string; showBar: boolean; showLabel: boolean }[]
    | undefined;
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
const renderValueLineChartLines = (
  lines: ValueLineChartLine[],
  gradientId?: string,
): ReactElement[] => {
  return lines.map((line) => {
    return (
      <path
        d={line.path}
        fill="none"
        key={line.key}
        stroke={gradientId ? `url(#${gradientId})` : line.color}
        strokeWidth="2"
      />
    );
  });
};

/** Builds hover content from the series values at one timestamp. */
const renderValueLineChartTooltip = (
  hoveredPoint: ChartHoverPoint | null,
  {
    stamps,
    lines,
    valuesByKey,
    formatValue,
  }: Pick<ValueLineChartSvgProps, "stamps" | "lines" | "valuesByKey" | "formatValue">,
): ReactElement | null => {
  if (hoveredPoint === null) return null;
  const hoveredStamp = stamps[hoveredPoint.index];
  if (hoveredStamp === undefined) return null;
  return (
    <MonitoringChartTooltip
      chartWidth={valueLineChartFrame.chartWidth}
      items={lines.flatMap((line) => {
        const value = valuesByKey[line.key]?.[hoveredPoint.index];
        if (value === undefined) return [];
        return [{ color: line.color, label: line.key, value: formatValue(value) }];
      })}
      label={formatChartStamp(hoveredStamp)}
      x={hoveredPoint.x}
    />
  );
};

/** Draws optional horizontal sensor guides against the current vertical domain. */
const renderValueLineChartGuides = (
  guideStops: ValueLineChartSvgProps["guideStops"],
  maxValue: number,
  formatValue: (value: number) => string,
): ReactElement[] =>
  (guideStops ?? [])
    .filter((stop) => stop.showBar || stop.showLabel)
    .map((stop) => {
      const y =
        valueLineChartFrame.topPadding +
        valueLineChartFrame.innerHeight * (1 - Math.max(0, Math.min(1, stop.value / maxValue)));
      return (
        <g key={`${stop.value}-${stop.color}-${stop.label}`}>
          {stop.showBar && (
            <line
              stroke={stop.color}
              strokeDasharray="6 4"
              x1={valueLineChartFrame.leftPadding}
              x2={valueLineChartFrame.chartWidth - valueLineChartFrame.rightPadding}
              y1={y}
              y2={y}
            />
          )}
          {stop.showLabel && (
            <text
              fill={stop.color}
              fontSize="11"
              textAnchor="end"
              x={valueLineChartFrame.chartWidth - valueLineChartFrame.rightPadding}
              y={y - 4}
            >
              {stop.label || formatValue(stop.value)}
            </text>
          )}
        </g>
      );
    });

/** Fills the area beneath each line when a color gradient is configured. */
const renderValueLineChartFills = (
  lines: ValueLineChartLine[],
  gradientId: string,
): Array<ReactElement | null> =>
  lines.map((line) => {
    const points = [...line.path.matchAll(/[ML] ([\d.]+) ([\d.]+)/g)];
    const first = points[0]?.[1];
    const last = points.at(-1)?.[1];
    if (!first || !last) return null;
    const baseline = valueLineChartFrame.topPadding + valueLineChartFrame.innerHeight;
    return (
      <path
        d={`${line.path} L ${last} ${baseline} L ${first} ${baseline} Z`}
        fill={`url(#${gradientId})`}
        key={`${line.key}-fill`}
        opacity="0.25"
      />
    );
  });

/** SVG frame rendered by the dynamic value line chart component. */
export const ValueLineChartSvg = ({
  axisLabels,
  capacity,
  colorStops,
  formatValue,
  guideStops,
  lines,
  stamps,
  valuesByKey,
  yAxisLabels,
}: ValueLineChartSvgProps): ReactElement => {
  const gradientId = useId().replaceAll(":", "");
  const maxValue = yAxisLabels[0]?.value ?? 1;
  const stops = [...(colorStops ?? [])].sort((a, b) => a.value - b.value);
  return (
    <ChartSvgContainer
      ariaLabel="Value line chart"
      capacity={capacity}
      frame={valueLineChartFrame}
      length={stamps.length}
      testId="value-line-chart"
      tooltip={(hoveredPoint) =>
        renderValueLineChartTooltip(hoveredPoint, { stamps, lines, valuesByKey, formatValue })
      }
    >
      {stops.length > 0 && (
        <defs>
          <linearGradient
            gradientUnits="userSpaceOnUse"
            id={gradientId}
            x1="0"
            x2="0"
            y1={valueLineChartFrame.topPadding + valueLineChartFrame.innerHeight}
            y2={valueLineChartFrame.topPadding}
          >
            {stops.map((stop) => (
              <stop
                key={`${stop.value}-${stop.color}`}
                offset={`${Math.max(0, Math.min(100, (stop.value / maxValue) * 100))}%`}
                stopColor={stop.color}
              />
            ))}
          </linearGradient>
        </defs>
      )}
      {renderValueLineChartGrid(yAxisLabels)}
      {renderValueLineChartGuides(guideStops, maxValue, formatValue)}
      {renderValueLineChartAxisLabels(axisLabels)}
      {stops.length > 0 && renderValueLineChartFills(lines, gradientId)}
      {renderValueLineChartLines(lines, stops.length > 0 ? gradientId : undefined)}
    </ChartSvgContainer>
  );
};
