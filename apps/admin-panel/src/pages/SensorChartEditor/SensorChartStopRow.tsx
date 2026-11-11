import { useLanguage } from "@lite-nas/language";
import Button from "@mui/material/Button";
import Checkbox from "@mui/material/Checkbox";
import TableCell from "@mui/material/TableCell";
import TableRow from "@mui/material/TableRow";
import TextField from "@mui/material/TextField";
import type { SensorChart } from "@schemas/monitoring/sensor";
import type { ReactElement } from "react";
import { SensorChartColorPicker } from "./SensorChartColorPicker";

type Stop = SensorChart["stops"][number];
type Props = {
  index: number;
  stop: Stop;
  onChange: (index: number, patch: Partial<Stop>) => void;
  onRemove: (index: number) => void;
};

/** Edits one stable chart stop row and reports changes to its owning dialog. */
export const SensorChartStopRow = ({ index, stop, onChange, onRemove }: Props): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <TableRow data-test-class="sensor-chart-stop" data-test-name={String(index)}>
      <TableCell>
        <TextField
          aria-label={getLabel("admin.sensors.value")}
          onChange={(event) => onChange(index, { value: Number(event.target.value) })}
          size="small"
          type="number"
          value={stop.value}
        />
      </TableCell>
      <TableCell>
        <SensorChartColorPicker
          color={stop.color}
          index={index}
          onChange={(color) => onChange(index, { color })}
        />
      </TableCell>
      <TableCell>
        <TextField
          aria-label={getLabel("admin.sensors.lineLabel")}
          onChange={(event) => onChange(index, { label: event.target.value })}
          size="small"
          value={stop.label}
        />
      </TableCell>
      <TableCell>
        <Checkbox
          aria-label={getLabel("admin.sensors.showBar")}
          checked={stop.showBar}
          onChange={(event) => onChange(index, { showBar: event.target.checked })}
        />
      </TableCell>
      <TableCell>
        <Checkbox
          aria-label={getLabel("admin.sensors.showLabel")}
          checked={stop.showLabel}
          onChange={(event) => onChange(index, { showLabel: event.target.checked })}
        />
      </TableCell>
      <TableCell>
        <Button onClick={() => onRemove(index)}>{getLabel("common.delete")}</Button>
      </TableCell>
    </TableRow>
  );
};
