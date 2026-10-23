import ExpandMoreIcon from "@mui/icons-material/ExpandMore";
import {
  Button,
  ClickAwayListener,
  LinearProgress,
  Paper,
  Popper,
  Stack,
  Typography,
} from "@mui/material";
import type { ReactElement } from "react";
import { useState } from "react";

import type { ExplorerQueuedAction } from "./FileExplorerContext";
import { useFileExplorer } from "./FileExplorerContext";

/** Shows the currently running action beneath total queue progress. */
function RunningActionProgress({ action }: { action: ExplorerQueuedAction }): ReactElement {
  return (
    <>
      <Typography variant="caption" noWrap>
        Current: {action.label} · {Math.round(action.progress)}%
      </Typography>
      <LinearProgress
        aria-label="Current operation progress"
        variant="determinate"
        value={action.progress}
      />
    </>
  );
}

/** Overall queue progress with expandable per-action results. */
export function ExplorerActionQueue(): ReactElement | null {
  const { queue, clearQueue } = useFileExplorer();
  const [anchor, setAnchor] = useState<HTMLElement | null>(null);
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
  const complete = completed === queue.length;
  const open = anchor !== null;

  return (
    <>
      <Button
        size="small"
        endIcon={<ExpandMoreIcon />}
        onClick={(event) => setAnchor(open ? null : event.currentTarget)}
        aria-expanded={open}
        aria-controls="file-explorer-action-queue"
      >
        Actions · {completed} of {queue.length} finished
      </Button>
      <Popper open={open} anchorEl={anchor} placement="bottom-end" sx={{ zIndex: 1300 }}>
        <ClickAwayListener onClickAway={() => setAnchor(null)}>
          <Paper
            id="file-explorer-action-queue"
            variant="outlined"
            data-testid="file-explorer-action-queue"
            sx={{ mt: 1, p: 2, width: { xs: "min(90vw, 380px)", sm: 380 }, boxShadow: 4 }}
          >
            <Stack spacing={1.5}>
              <Stack spacing={0.75}>
                <Typography variant="subtitle2">Action progress</Typography>
                <LinearProgress variant="determinate" value={overallProgress} />
                {running && <RunningActionProgress action={running} />}
              </Stack>
              <Stack spacing={0.25}>
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
              {complete && (
                <Button
                  size="small"
                  onClick={() => {
                    setAnchor(null);
                    clearQueue();
                  }}
                >
                  Close completed actions
                </Button>
              )}
            </Stack>
          </Paper>
        </ClickAwayListener>
      </Popper>
    </>
  );
}
