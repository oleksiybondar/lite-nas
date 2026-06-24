import {
  buildSortedNumericAlertOccurrences,
  formatAlertOccurrenceValue,
} from "@components/alerts/AlertOccurrencesDialog/helpers";
import { ChartSvgContainer } from "@components/monitoring/ChartSvgContainer";
import { MonitoringChartTooltip } from "@components/monitoring/MonitoringChartTooltip";
import {
  buildEvenlyDistributedIndexes,
  createPercentChartFrame,
  formatChartStamp,
  mapPercentChartSeriesX,
} from "@components/monitoring/percent-chart-shared";
import type { AlertOccurrenceItemDTO } from "@dto/alerts/alerts";
import { useTheme } from "@mui/material/styles";
import Typography from "@mui/material/Typography";
import { type ReactElement, useMemo } from "react";
import type { AlertOccurrencesChartProps } from "./types";

const yAxisSegments = 5;
const minChartCapacity = 2;
const maxChartCapacity = 10000;
const chartEntriesPerLabel = 80;
const chartFrame = createPercentChartFrame({
  bottomPadding: 34,
  chartHeight: 320,
  chartWidth: 760,
  leftPadding: 56,
  rightPadding: 20,
  topPadding: 12,
});

/**
 * Numeric chart view for alert occurrences, including negative-value support.
 */
export const AlertOccurrencesChart = ({
  occurrences,
}: AlertOccurrencesChartProps): ReactElement => {
  const theme = useTheme();
  const chartState = useAlertOccurrencesChartState(occurrences);

  if (chartState === null) {
    return renderAlertOccurrencesChartEmptyState();
  }

  return (
    <ChartSvgContainer
      ariaLabel="Alert occurrences chart"
      capacity={chartState.chartCapacity}
      frame={chartFrame}
      heightPx={320}
      length={chartState.chartPoints.length}
      testId="alert-occurrences-chart"
      tooltip={buildAlertOccurrencesTooltipRenderer(chartState, theme.palette.primary.main)}
    >
      {renderAlertOccurrencesGrid(theme.palette.divider, chartState.yAxisLabels)}
      {renderAlertOccurrencesZeroLineIfNeeded(theme.palette.text.primary, chartState.zeroLineY)}
      {renderAlertOccurrencesAxisLabels(chartState.axisLabels, theme.palette.text.secondary)}
      <path
        d={chartState.linePath}
        fill="none"
        stroke={theme.palette.primary.main}
        strokeWidth="2.5"
      />
    </ChartSvgContainer>
  );
};

type AlertOccurrencesAxisLabel = {
  label: string;
  x: number;
};

type AlertOccurrencesChartPoint = {
  occurrence: AlertOccurrenceItemDTO | null;
  value: number;
};

type AlertOccurrencesYAxisLabel = {
  label: string;
  value: number;
  y: number;
};

type AlertOccurrencesChartState = {
  axisLabels: AlertOccurrencesAxisLabel[];
  chartCapacity: number;
  chartPoints: AlertOccurrencesChartPoint[];
  linePath: string;
  yAxisLabels: AlertOccurrencesYAxisLabel[];
  zeroLineY: number | null;
};

/**
 * Resolves the prepared chart state for numeric alert occurrences.
 */
const useAlertOccurrencesChartState = (
  occurrences: AlertOccurrenceItemDTO[],
): AlertOccurrencesChartState | null => {
  const sortedOccurrences = useMemo(() => {
    return buildSortedNumericAlertOccurrences(occurrences);
  }, [occurrences]);
  const numericOccurrences = sortedOccurrences.filter((occurrence) => occurrence.ValueNum !== null);

  if (numericOccurrences.length === 0) {
    return null;
  }

  const chartPoints = buildAlertOccurrencesChartPoints(numericOccurrences);
  const chartCapacity = resolveAlertOccurrencesChartCapacity(chartPoints.length);
  const yDomain = buildAlertOccurrencesYDomain(chartPoints);

  return {
    axisLabels: buildAlertOccurrencesAxisLabels(chartPoints, chartCapacity),
    chartCapacity,
    chartPoints,
    linePath: buildAlertOccurrencesLinePath(chartPoints, chartCapacity, yDomain.min, yDomain.max),
    yAxisLabels: buildAlertOccurrencesYAxisLabels(yDomain.min, yDomain.max),
    zeroLineY: resolveAlertOccurrencesZeroLineY(yDomain.min, yDomain.max),
  };
};

/**
 * Renders the chart empty state when no numeric occurrences are available.
 */
const renderAlertOccurrencesChartEmptyState = (): ReactElement => {
  return (
    <Typography color="text.secondary" data-testid="alert-occurrences-chart-empty" variant="body2">
      Chart data will appear after numeric occurrences are loaded.
    </Typography>
  );
};

/**
 * Builds evenly distributed X-axis labels for the current numeric occurrences.
 */
const buildAlertOccurrencesAxisLabels = (
  points: AlertOccurrencesChartPoint[],
  capacity: number,
): AlertOccurrencesAxisLabel[] => {
  const visiblePoints = points.filter(isAlertOccurrencesDataPoint);

  if (visiblePoints.length === 0) {
    return [];
  }

  const labelCount = Math.min(
    visiblePoints.length,
    Math.max(1, Math.ceil(visiblePoints.length / chartEntriesPerLabel)),
  );
  const labelIndexes = buildEvenlyDistributedIndexes(visiblePoints.length, labelCount);

  return labelIndexes.flatMap((index) => {
    const occurrence = visiblePoints[index]?.occurrence;

    if (occurrence === undefined) {
      return [];
    }

    return [
      {
        label: formatChartStamp(occurrence.Timestamp),
        x: mapPercentChartSeriesX(chartFrame, index + 1, points.length, capacity),
      },
    ];
  });
};

/**
 * Builds the visible Y-axis labels for the current numeric value domain.
 */
const buildAlertOccurrencesYAxisLabels = (
  minValue: number,
  maxValue: number,
): AlertOccurrencesYAxisLabel[] => {
  const range = maxValue - minValue;

  return Array.from({ length: yAxisSegments + 1 }, (_, index) => {
    const ratio = (yAxisSegments - index) / yAxisSegments;
    const value = minValue + range * ratio;

    return {
      label: formatAlertOccurrencesAxisValue(value),
      value,
      y: mapAlertOccurrencesY(value, minValue, maxValue),
    };
  });
};

/**
 * Builds the SVG line path for the numeric occurrences series.
 */
const buildAlertOccurrencesLinePath = (
  points: AlertOccurrencesChartPoint[],
  capacity: number,
  minValue: number,
  maxValue: number,
): string => {
  return points
    .map((point, index) => {
      const x = mapPercentChartSeriesX(chartFrame, index, points.length, capacity);
      const y = mapAlertOccurrencesY(point.value, minValue, maxValue);
      const command = index === 0 ? "M" : "L";

      return `${command} ${x} ${y}`;
    })
    .join(" ");
};

/**
 * Resolves the numeric value domain, including a stable padded range around zero-only data.
 */
const buildAlertOccurrencesYDomain = (
  points: AlertOccurrencesChartPoint[],
): {
  max: number;
  min: number;
} => {
  const values = points.map((point) => point.value);
  const rawMin = Math.min(...values);
  const rawMax = Math.max(...values);

  if (rawMin === rawMax) {
    if (rawMin === 0) {
      return { max: 1, min: -1 };
    }

    const offset = Math.max(1, Math.abs(rawMin) * 0.1);
    return {
      max: rawMax + offset,
      min: rawMin - offset,
    };
  }

  const paddedMin = padAlertOccurrencesDomain(rawMin, rawMax, "min");
  const paddedMax = padAlertOccurrencesDomain(rawMin, rawMax, "max");

  return {
    max: roundAlertOccurrencesDomainValue(paddedMax, "max"),
    min: roundAlertOccurrencesDomainValue(paddedMin, "min"),
  };
};

/**
 * Builds the plotted chart-point list, including a synthetic first point that
 * repeats the first occurrence value so the chart starts with a horizontal segment.
 */
const buildAlertOccurrencesChartPoints = (
  occurrences: AlertOccurrenceItemDTO[],
): AlertOccurrencesChartPoint[] => {
  const firstValue = occurrences[0]?.ValueNum ?? 0;

  return [
    {
      occurrence: null,
      value: firstValue,
    },
    ...occurrences.map((occurrence) => {
      return {
        occurrence,
        value: occurrence.ValueNum ?? 0,
      };
    }),
  ];
};

/**
 * Narrows one plotted chart point to a real occurrence-backed data point.
 */
const isAlertOccurrencesDataPoint = (
  point: AlertOccurrencesChartPoint,
): point is AlertOccurrencesChartPoint & { occurrence: AlertOccurrenceItemDTO } => {
  return point.occurrence !== null;
};

/**
 * Resolves the chart capacity from the current point count instead of using a fixed slot window.
 */
const resolveAlertOccurrencesChartCapacity = (pointCount: number): number => {
  return Math.max(minChartCapacity, Math.min(maxChartCapacity, pointCount));
};

/**
 * Maps one numeric value to the chart Y coordinate for the current domain.
 */
const mapAlertOccurrencesY = (value: number, minValue: number, maxValue: number): number => {
  const safeRange = Math.max(maxValue - minValue, Number.EPSILON);
  const boundedValue = Math.min(maxValue, Math.max(minValue, value));
  const ratio = (boundedValue - minValue) / safeRange;

  return chartFrame.topPadding + chartFrame.innerHeight - ratio * chartFrame.innerHeight;
};

/**
 * Resolves the zero-line Y coordinate when zero falls inside the visible domain.
 */
const resolveAlertOccurrencesZeroLineY = (minValue: number, maxValue: number): number | null => {
  if (minValue > 0 || maxValue < 0) {
    return null;
  }

  return mapAlertOccurrencesY(0, minValue, maxValue);
};

/**
 * Builds the hover-tooltip renderer for the numeric occurrences chart.
 */
const buildAlertOccurrencesTooltipRenderer = (
  chartState: AlertOccurrencesChartState,
  lineColor: string,
): ((hoveredPoint: { index: number; x: number } | null) => ReactElement | null) => {
  return (hoveredPoint) => {
    if (hoveredPoint === null) {
      return null;
    }

    const occurrence = chartState.chartPoints[hoveredPoint.index]?.occurrence;

    if (occurrence === null || occurrence === undefined) {
      return null;
    }

    return (
      <MonitoringChartTooltip
        chartWidth={chartFrame.chartWidth}
        items={[
          {
            color: lineColor,
            label: buildAlertOccurrencesTooltipLabel(occurrence),
            value: formatAlertOccurrenceValue(occurrence),
          },
        ]}
        label={formatChartStamp(occurrence.Timestamp)}
        x={hoveredPoint.x}
      />
    );
  };
};

/**
 * Resolves the tooltip series label for one numeric occurrence.
 */
const buildAlertOccurrencesTooltipLabel = (occurrence: AlertOccurrenceItemDTO): string => {
  if (occurrence.ValueUnit === null || occurrence.ValueUnit === "") {
    return "Value";
  }

  return `Value (${occurrence.ValueUnit})`;
};

/**
 * Renders the horizontal grid lines and Y-axis labels for the numeric chart.
 */
const renderAlertOccurrencesGrid = (
  lineColor: string,
  labels: AlertOccurrencesYAxisLabel[],
): ReactElement[] => {
  return labels.map((label, index) => {
    return (
      <g key={`${label.value}-${label.label}`}>
        <line
          stroke={lineColor}
          strokeDasharray={index === labels.length - 1 ? undefined : "4 6"}
          strokeOpacity={0.5}
          strokeWidth="1"
          x1={chartFrame.leftPadding}
          x2={chartFrame.chartWidth - chartFrame.rightPadding}
          y1={label.y}
          y2={label.y}
        />
        <text
          fill="currentColor"
          fontSize="12"
          textAnchor="end"
          x={chartFrame.leftPadding - 8}
          y={label.y + 4}
        >
          {label.label}
        </text>
      </g>
    );
  });
};

/**
 * Renders the emphasized horizontal zero line when the value domain crosses zero.
 */
const renderAlertOccurrencesZeroLine = (lineColor: string, y: number): ReactElement => {
  return (
    <line
      stroke={lineColor}
      strokeOpacity={0.7}
      strokeWidth="1.5"
      x1={chartFrame.leftPadding}
      x2={chartFrame.chartWidth - chartFrame.rightPadding}
      y1={y}
      y2={y}
    />
  );
};

/**
 * Renders the zero line only when the current value domain crosses zero.
 */
const renderAlertOccurrencesZeroLineIfNeeded = (
  lineColor: string,
  y: number | null,
): ReactElement | null => {
  if (y === null) {
    return null;
  }

  return renderAlertOccurrencesZeroLine(lineColor, y);
};

/**
 * Renders the X-axis timestamp labels for the numeric chart.
 */
const renderAlertOccurrencesAxisLabels = (
  labels: AlertOccurrencesAxisLabel[],
  textColor: string,
): ReactElement[] => {
  return labels.map((label) => {
    return (
      <text
        fill={textColor}
        fontSize="12"
        key={`${label.x}-${label.label}`}
        textAnchor="middle"
        x={label.x}
        y={chartFrame.chartHeight - 8}
      >
        {label.label}
      </text>
    );
  });
};

/**
 * Formats one Y-axis numeric value using a compact human-readable representation.
 */
const formatAlertOccurrencesAxisValue = (value: number): string => {
  return new Intl.NumberFormat(undefined, {
    maximumFractionDigits: Math.abs(value) >= 10 ? 0 : 2,
    notation: Math.abs(value) >= 1000 ? "compact" : "standard",
  }).format(value);
};

/**
 * Pads one side of the visible numeric domain so points do not hug the frame edges.
 */
const padAlertOccurrencesDomain = (
  minValue: number,
  maxValue: number,
  side: "max" | "min",
): number => {
  const range = maxValue - minValue;
  const padding = Math.max(range * 0.05, Number.EPSILON);

  return side === "min" ? minValue - padding : maxValue + padding;
};

/**
 * Rounds one padded domain bound outward to a stable human-readable value.
 */
const roundAlertOccurrencesDomainValue = (value: number, side: "max" | "min"): number => {
  if (value === 0) {
    return 0;
  }

  const magnitude = 10 ** Math.floor(Math.log10(Math.abs(value)));
  const scaled = value / magnitude;

  if (side === "max") {
    return Math.ceil(scaled) * magnitude;
  }

  return Math.floor(scaled) * magnitude;
};
