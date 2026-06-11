import Box from "@mui/material/Box";
import type { MouseEvent, ReactElement, ReactNode } from "react";
import { useState } from "react";
import {
  type ChartHoverPoint,
  type PercentChartFrame,
  resolveHoveredChartPointFromClientX,
} from "./percent-chart-shared";

type ChartSvgContainerProps = {
  ariaLabel: string;
  capacity: number;
  children: ReactNode;
  frame: PercentChartFrame;
  heightPx?: number | undefined;
  length: number;
  testId: string;
  tooltip: (hoveredPoint: ChartHoverPoint | null) => ReactNode;
};

/**
 * Shared wrapper that owns hover-state management and the Box/svg scaffold for monitoring charts.
 */
export const ChartSvgContainer = ({
  ariaLabel,
  capacity,
  children,
  frame,
  heightPx,
  length,
  testId,
  tooltip,
}: ChartSvgContainerProps): ReactElement => {
  const [hoveredPoint, setHoveredPoint] = useState<ChartHoverPoint | null>(null);

  const handleMouseLeave = (): void => {
    setHoveredPoint(null);
  };

  const handleMouseMove = (event: MouseEvent<HTMLDivElement>): void => {
    const rect = event.currentTarget.getBoundingClientRect();

    setHoveredPoint(
      resolveHoveredChartPointFromClientX({
        capacity,
        clientX: event.clientX,
        frame,
        length,
        rectLeft: rect.left,
        rectWidth: rect.width,
      }),
    );
  };

  return (
    <Box
      data-testid={testId}
      onMouseLeave={handleMouseLeave}
      onMouseMove={handleMouseMove}
      sx={{ height: heightPx, overflow: "visible", position: "relative", width: "100%" }}
    >
      <svg
        aria-label={ariaLabel}
        height={heightPx}
        preserveAspectRatio="none"
        viewBox={`0 0 ${frame.chartWidth} ${frame.chartHeight}`}
        width="100%"
      >
        {children}
      </svg>
      {tooltip(hoveredPoint)}
    </Box>
  );
};
