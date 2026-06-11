import { formatFilterSelection } from "@components/alerts/AlertsControlPanel/helpers";
import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import MenuItem from "@mui/material/MenuItem";
import type { SelectChangeEvent } from "@mui/material/Select";
import Select from "@mui/material/Select";
import type { ReactElement } from "react";

/**
 * One selectable option rendered by the reusable multiselect filter.
 */
type MultiValueFilterOption = {
  label: string;
  value: number | string;
};

/**
 * Props accepted by the reusable alerts multiselect filter input.
 */
type AlertsControlPanelMultiValueFilterProps = {
  label: string;
  name: string;
  onChange: (value: string[]) => void;
  options: MultiValueFilterOption[];
  value: string[] | number[];
};

/**
 * Renders one reusable multi-select filter input for the alerts control panel.
 */
export const AlertsControlPanelMultiValueFilter = ({
  label,
  name,
  onChange,
  options,
  value,
}: AlertsControlPanelMultiValueFilterProps): ReactElement => {
  const labelId = `${name}-label`;

  return (
    <FormControl sx={{ width: 220, maxWidth: "100%" }}>
      <InputLabel id={labelId}>{label}</InputLabel>
      <Select
        data-testid={`${name}-select`}
        label={label}
        labelId={labelId}
        multiple
        name={name}
        onChange={(event) => {
          onChange(readSelectValues(event));
        }}
        renderValue={(selected) => {
          return formatFilterSelection(selected as string[]);
        }}
        value={value.map(String)}
      >
        {options.map((option) => {
          return (
            <MenuItem
              data-test-class="alerts-filter-option"
              data-test-name={`${label}:${option.label}`}
              key={`${name}-${option.value}`}
              value={String(option.value)}
            >
              {option.label}
            </MenuItem>
          );
        })}
      </Select>
    </FormControl>
  );
};

/**
 * Normalizes MUI select output into a stable string-array value contract.
 */
const readSelectValues = (event: SelectChangeEvent<string[]>): string[] => {
  const nextValue = event.target.value;

  return typeof nextValue === "string" ? nextValue.split(",") : nextValue;
};
