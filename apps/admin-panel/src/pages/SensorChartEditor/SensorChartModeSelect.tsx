import { useLanguage } from "@lite-nas/language";
import FormControlLabel from "@mui/material/FormControlLabel";
import Radio from "@mui/material/Radio";
import RadioGroup from "@mui/material/RadioGroup";
import type { SensorChart } from "@schemas/monitoring/sensor";
import type { ReactElement } from "react";

/** Selects the regular or gradient presentation for one sensor chart. */
export const SensorChartModeSelect = ({
  mode,
  onChange,
}: {
  mode: SensorChart["mode"];
  onChange: (mode: SensorChart["mode"]) => void;
}): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <RadioGroup
      onChange={(event) => onChange(event.target.value as SensorChart["mode"])}
      row
      value={mode}
    >
      <FormControlLabel
        control={<Radio />}
        label={getLabel("admin.sensors.regularChart")}
        value="regular"
      />
      <FormControlLabel
        control={<Radio />}
        label={getLabel("admin.sensors.gradientChart")}
        value="gradient"
      />
    </RadioGroup>
  );
};
