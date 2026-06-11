import Box from "@mui/material/Box";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import type { MonitoringChartTooltipProps } from "./types";

/**
 * Floating hover surface shared by monitoring charts to show one timestamp and its series values.
 */
export const MonitoringChartTooltip = ({
  chartWidth,
  items,
  label,
  x,
}: MonitoringChartTooltipProps): ReactElement => {
  const anchorPercent = (x / chartWidth) * 100;
  const isRightHalf = x > chartWidth / 2;

  return (
    <Paper
      data-testid="monitoring-chart-tooltip"
      elevation={6}
      sx={{
        left: `${anchorPercent}%`,
        maxWidth: 220,
        p: 1.25,
        pointerEvents: "none",
        position: "absolute",
        top: 8,
        transform: isRightHalf ? "translateX(calc(-100% - 8px))" : "translateX(8px)",
        zIndex: 2,
      }}
    >
      <Stack spacing={0.75}>
        <Typography variant="caption">{label}</Typography>
        <Stack spacing={0.5}>
          {items.map((item) => {
            return (
              <Stack alignItems="center" direction="row" key={item.label} spacing={0.75}>
                {item.color !== undefined ? (
                  <Box
                    sx={{
                      backgroundColor: item.color,
                      borderRadius: "999px",
                      flexShrink: 0,
                      height: 8,
                      width: 8,
                    }}
                  />
                ) : null}
                <Typography variant="body2">
                  {item.label}: {item.value}
                </Typography>
              </Stack>
            );
          })}
        </Stack>
      </Stack>
    </Paper>
  );
};
