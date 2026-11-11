import { useLanguage } from "@lite-nas/language";
import Button from "@mui/material/Button";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogTitle from "@mui/material/DialogTitle";
import Stack from "@mui/material/Stack";
import Table from "@mui/material/Table";
import TableBody from "@mui/material/TableBody";
import TableCell from "@mui/material/TableCell";
import TableContainer from "@mui/material/TableContainer";
import TableHead from "@mui/material/TableHead";
import TableRow from "@mui/material/TableRow";
import Typography from "@mui/material/Typography";
import { SensorChartModeSelect, SensorChartStopRow } from "@pages/SensorChartEditor/index";
import { type SensorChart, sensorChartSchema } from "@schemas/monitoring/sensor";
import { type Dispatch, memo, type ReactElement, type SetStateAction, useState } from "react";

/** Information supplied by the sensor that opened the editor. */
type Props = {
  chart: SensorChart;
  name: string;
  onClose: () => void;
  onSave: (chart: SensorChart) => Promise<void>;
};

/** Applies one edited stop without replacing the other stops. */
const updateStop = (
  setDraft: Dispatch<SetStateAction<SensorChart>>,
  index: number,
  patch: Partial<SensorChart["stops"][number]>,
): void => {
  setDraft((current) => ({
    ...current,
    stops: current.stops.map((stop, at) => (at === index ? { ...stop, ...patch } : stop)),
  }));
};

/** Removes a stop and its matching stable React row identifier. */
const removeStop = (
  setDraft: Dispatch<SetStateAction<SensorChart>>,
  setRowIds: Dispatch<SetStateAction<string[]>>,
  index: number,
): void => {
  setRowIds((current) => current.filter((_, at) => at !== index));
  setDraft((current) => ({ ...current, stops: current.stops.filter((_, at) => at !== index) }));
};

/** Validates and saves the editor draft, reporting any failed attempt. */
const saveDraft = async ({
  draft,
  onSave,
  onClose,
  setError,
  setSaving,
}: {
  draft: SensorChart;
  onSave: Props["onSave"];
  onClose: Props["onClose"];
  setError: (error: boolean) => void;
  setSaving: (saving: boolean) => void;
}): Promise<void> => {
  const parsed = sensorChartSchema.safeParse(draft);
  if (!parsed.success) {
    setError(true);
    return;
  }
  setSaving(true);
  try {
    await onSave(parsed.data);
    onClose();
  } catch {
    setError(true);
  } finally {
    setSaving(false);
  }
};

/** Owns one sensor's draft and saves its validated chart JSON. */
const SensorChartEditorContent = ({ chart, name, onClose, onSave }: Props): ReactElement => {
  const { getLabel } = useLanguage();
  const [draft, setDraft] = useState(chart);
  const [rowIds, setRowIds] = useState<string[]>(() => chart.stops.map(() => crypto.randomUUID()));
  const [error, setError] = useState(false);
  const [saving, setSaving] = useState(false);
  return (
    <Dialog data-testid="sensor-chart-editor" fullWidth maxWidth="md" onClose={onClose} open>
      <DialogTitle>
        {getLabel("admin.sensors.editChart")}: {name}
      </DialogTitle>
      <DialogContent>
        <Stack spacing={2} sx={{ pt: 1 }}>
          <SensorChartModeSelect
            mode={draft.mode}
            onChange={(mode) => setDraft((current) => ({ ...current, mode }))}
          />
          <TableContainer>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>{getLabel("admin.sensors.value")}</TableCell>
                  <TableCell>{getLabel("admin.sensors.color")}</TableCell>
                  <TableCell>{getLabel("admin.sensors.lineLabel")}</TableCell>
                  <TableCell>{getLabel("admin.sensors.showBar")}</TableCell>
                  <TableCell>{getLabel("admin.sensors.showLabel")}</TableCell>
                  <TableCell />
                </TableRow>
              </TableHead>
              <TableBody>
                {draft.stops
                  .map((stop, index) => ({ stop, index, id: rowIds[index] }))
                  .map(({ stop, index, id }) => (
                    <SensorChartStopRow
                      key={id}
                      index={index}
                      stop={stop}
                      onChange={(index, patch) => updateStop(setDraft, index, patch)}
                      onRemove={(index) => removeStop(setDraft, setRowIds, index)}
                    />
                  ))}
              </TableBody>
            </Table>
          </TableContainer>
          <Button
            data-testid="sensor-chart-add-stop"
            disabled={draft.stops.length >= 32}
            onClick={() => {
              setRowIds((current) => [...current, crypto.randomUUID()]);
              setDraft((current) => ({
                ...current,
                stops: [
                  ...current.stops,
                  { value: 0, color: "#00bfff", label: "", showBar: false, showLabel: false },
                ],
              }));
            }}
          >
            {getLabel("admin.sensors.addStop")}
          </Button>
          {error && (
            <Typography color="error" role="alert">
              {getLabel("admin.sensors.invalidMarkers")}
            </Typography>
          )}
        </Stack>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>{getLabel("common.cancel")}</Button>
        <Button
          data-testid="sensor-chart-save"
          disabled={saving}
          onClick={() => void saveDraft({ draft, onSave, onClose, setError, setSaving })}
        >
          {getLabel("common.save")}
        </Button>
      </DialogActions>
    </Dialog>
  );
};

/** Keeps the open editor isolated from one-second sensor page renders. */
export const SensorChartEditor = memo(
  SensorChartEditorContent,
  (previous, next) => previous.name === next.name,
);
