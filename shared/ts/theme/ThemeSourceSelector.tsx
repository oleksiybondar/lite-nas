import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import type { ReactElement } from "react";
import type { ThemeSource } from "./types";

/** Controlled source selector shared by applications using LiteNAS themes. */
export const ThemeSourceSelector = ({
  onChange,
  value,
}: {
  onChange: (source: ThemeSource) => void;
  value: ThemeSource;
}): ReactElement => (
  <FormControl data-testid="theme-source-control" fullWidth>
    <InputLabel data-testid="theme-source-label" id="theme-source-label">
      Theme source
    </InputLabel>
    <Select
      data-testid="theme-source-select"
      label="Theme source"
      labelId="theme-source-label"
      name="themeSource"
      onChange={(event) => onChange(event.target.value as ThemeSource)}
      value={value}
    >
      <MenuItem data-test-class="theme-source-option" data-test-name="Default" value="default">
        Default
      </MenuItem>
      <MenuItem data-test-class="theme-source-option" data-test-name="OS" value="os">
        OS
      </MenuItem>
      <MenuItem data-test-class="theme-source-option" data-test-name="User" value="user">
        User
      </MenuItem>
    </Select>
  </FormControl>
);
