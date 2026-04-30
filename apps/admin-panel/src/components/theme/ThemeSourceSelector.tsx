import { useThemeManager } from "@hooks/useThemeManager";

import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import type { ThemeSource } from "@theme/index";
import type { ReactElement } from "react";

interface ThemeSourceSelectorProps {
  onChange?: (source: ThemeSource) => void;
  value?: ThemeSource;
}

export const ThemeSourceSelector = ({
  onChange,
  value,
}: ThemeSourceSelectorProps): ReactElement => {
  const { setSource, source } = useThemeManager();
  const selectedValue = value ?? source;

  return (
    <FormControl fullWidth>
      <InputLabel id="theme-source-label">Theme source</InputLabel>
      <Select
        label="Theme source"
        labelId="theme-source-label"
        onChange={(event) => {
          const nextValue = event.target.value as ThemeSource;

          if (onChange !== undefined) {
            onChange(nextValue);
            return;
          }

          setSource(nextValue);
        }}
        value={selectedValue}
      >
        <MenuItem value="default">Default</MenuItem>
        <MenuItem value="os">OS</MenuItem>
        <MenuItem value="user">User</MenuItem>
      </Select>
    </FormControl>
  );
};
