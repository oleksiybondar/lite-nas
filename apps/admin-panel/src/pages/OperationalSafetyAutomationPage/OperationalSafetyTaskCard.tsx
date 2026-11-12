import { useLanguage } from "@lite-nas/language";
import {
  Button,
  FormControl,
  InputLabel,
  MenuItem,
  Paper,
  Select,
  Stack,
  Typography,
} from "@mui/material";
import type { OperationalSafetyTask } from "@schemas/operational-safety-automation";
import { type ReactElement, useState } from "react";

const levels: OperationalSafetyTask["level"][] = ["normal", "warning", "error", "critical"];

/** Displays one active task and its validated operator controls. */
export const OperationalSafetyTaskCard = ({
  task,
  onForceSet,
  onForceStop,
}: {
  task: OperationalSafetyTask;
  onForceSet: (id: string, level: OperationalSafetyTask["level"]) => void;
  onForceStop: (id: string) => void;
}): ReactElement => {
  const { getLabel } = useLanguage();
  const [selectedLevel, setSelectedLevel] = useState<OperationalSafetyTask["level"] | null>(null);
  return (
    <Paper
      variant="outlined"
      sx={{ p: 2 }}
      data-test-class="operational-safety-task"
      data-test-name={task.id}
    >
      <Stack spacing={1}>
        <Typography variant="h6">
          {task.rule_id} · {task.level}
        </Typography>
        <Typography variant="body2">
          {getLabel("admin.operationalSafety.taskId")}: {task.id} ·{" "}
          {getLabel("admin.operationalSafety.correlationKey")}: {task.correlation_key}
        </Typography>
        <Typography variant="body2">
          {getLabel("admin.operationalSafety.activeAlerts")}:{" "}
          {task.active_alert_ids.join(", ") || "—"} ·{" "}
          {getLabel("admin.operationalSafety.lastOutcome")}: {task.last_outcome}
        </Typography>
        {task.cooldown_until && (
          <Typography variant="body2">
            {getLabel("admin.operationalSafety.cooldownUntil")}: {task.cooldown_until}
          </Typography>
        )}
        <Stack direction={{ xs: "column", sm: "row" }} spacing={1}>
          <FormControl size="small" sx={{ minWidth: 150 }}>
            <InputLabel id={`level-label-${task.id}`}>
              {getLabel("admin.operationalSafety.level")}
            </InputLabel>
            <Select
              labelId={`level-label-${task.id}`}
              label={getLabel("admin.operationalSafety.level")}
              value={selectedLevel ?? task.level}
              onChange={(event) =>
                setSelectedLevel(event.target.value as OperationalSafetyTask["level"])
              }
            >
              {levels.map((level) => (
                <MenuItem key={level} value={level}>
                  {level}
                </MenuItem>
              ))}
            </Select>
          </FormControl>
          <Button onClick={() => onForceSet(task.id, selectedLevel ?? task.level)}>
            {getLabel("admin.operationalSafety.applyLevel")}
          </Button>
          <Button color="warning" onClick={() => onForceStop(task.id)}>
            {getLabel("admin.operationalSafety.forceStop")}
          </Button>
        </Stack>
      </Stack>
    </Paper>
  );
};
