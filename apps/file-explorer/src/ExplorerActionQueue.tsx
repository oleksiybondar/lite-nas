import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import {
  Accordion,
  AccordionDetails,
  AccordionSummary,
  LinearProgress,
  Stack,
  Typography,
} from "@mui/material";
import type { ReactElement } from "react";

import { useFileExplorer } from "./FileExplorerContext";

/** Overall queue progress with expandable per-action results. */
export function ExplorerActionQueue(): ReactElement | null {
  const { queue } = useFileExplorer();
  if (queue.length === 0) {
    return null;
  }
  const completed = queue.filter(
    (item) => item.status === "done" || item.status === "error",
  ).length;
  const running = queue.find((item) => item.status === "running");
  const overallProgress =
    queue.reduce((total, item) => total + (item.status === "error" ? 100 : item.progress), 0) /
    queue.length;

  return (
    <Accordion data-testid="file-explorer-action-queue">
      <AccordionSummary expandIcon={<ExpandMoreIcon />}>
        <Stack spacing={1} sx={{ width: "100%", pr: 2 }}>
          <Typography variant="subtitle2">
            Actions · {completed} of {queue.length} finished
          </Typography>
          <LinearProgress variant="determinate" value={overallProgress} />
          {running && (
            <Typography variant="caption">
              Current: {running.label} · {Math.round(running.progress)}%
            </Typography>
          )}
          {running && (
            <LinearProgress
              aria-label="Current operation progress"
              variant="determinate"
              value={running.progress}
            />
          )}
        </Stack>
      </AccordionSummary>
      <AccordionDetails>
        <Stack spacing={0.5}>
          {queue.map((item) => (
            <Typography
              key={item.id}
              variant="body2"
              color={item.status === "error" ? "error" : "text.secondary"}
            >
              {item.status}: {item.label} · {Math.round(item.progress)}%
              {item.error ? ` — ${item.error}` : ""}
            </Typography>
          ))}
        </Stack>
      </AccordionDetails>
    </Accordion>
  );
}
