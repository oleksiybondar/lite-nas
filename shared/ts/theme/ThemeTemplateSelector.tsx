import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import type { ReactElement } from "react";
import type { ThemeSource, ThemeTemplateName } from "./types";

/** Controlled visual template selector shared by LiteNAS settings pages. */
export const ThemeTemplateSelector = ({
  availableTemplates,
  onChange,
  source,
  value,
}: {
  availableTemplates: ThemeTemplateName[];
  onChange: (name: ThemeTemplateName) => void;
  source: ThemeSource;
  value: ThemeTemplateName;
}): ReactElement => (
  <FormControl data-testid="theme-template-control" disabled={source !== "user"} fullWidth>
    <InputLabel data-testid="theme-template-label" id="theme-template-label">
      Theme template
    </InputLabel>
    <Select
      data-testid="theme-template-select"
      label="Theme template"
      labelId="theme-template-label"
      name="themeTemplate"
      onChange={(event) => onChange(event.target.value as ThemeTemplateName)}
      value={value}
    >
      {availableTemplates.map((name) => {
        const label = name.slice(0, 1).toUpperCase() + name.slice(1);
        return (
          <MenuItem
            data-test-class="theme-template-option"
            data-test-name={label}
            key={name}
            value={name}
          >
            {label}
          </MenuItem>
        );
      })}
    </Select>
  </FormControl>
);
