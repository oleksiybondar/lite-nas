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
    <FormControl data-testid="theme-source-control" fullWidth>
      <InputLabel data-testid="theme-source-label" id="theme-source-label">
        Theme source
      </InputLabel>
      <Select
        data-testid="theme-source-select"
        label="Theme source"
        labelId="theme-source-label"
        name="themeSource"
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
};
