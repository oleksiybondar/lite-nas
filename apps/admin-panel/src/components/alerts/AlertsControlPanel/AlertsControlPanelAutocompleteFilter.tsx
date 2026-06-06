import type { AlertsControlPanelOption } from "@dto/alerts/alerts";
import Autocomplete from "@mui/material/Autocomplete";
import TextField from "@mui/material/TextField";
import type { ReactElement } from "react";

/**
 * Props accepted by the alerts autocomplete multi-filter input.
 */
type AlertsControlPanelAutocompleteFilterProps = {
  label: string;
  name: string;
  onChange: (value: string[]) => void;
  options: AlertsControlPanelOption<string>[];
  value: string[];
};

/**
 * Renders one multi-select autocomplete filter with custom-value support.
 */
export const AlertsControlPanelAutocompleteFilter = ({
  label,
  name,
  onChange,
  options,
  value,
}: AlertsControlPanelAutocompleteFilterProps): ReactElement => {
  return (
    <Autocomplete
      data-testid={`${name}-autocomplete`}
      freeSolo
      multiple
      onChange={(_, nextValue) => {
        onChange(nextValue);
      }}
      options={options.map((option) => option.value)}
      sx={{ width: 220, maxWidth: "100%" }}
      renderInput={(params) => {
        return <TextField {...params} label={label} name={name} />;
      }}
      value={value}
    />
  );
};
