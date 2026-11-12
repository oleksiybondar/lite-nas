import { useLanguage } from "@lite-nas/language";
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
}): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <FormControl data-testid="theme-template-control" disabled={source !== "user"} fullWidth>
      <InputLabel data-testid="theme-template-label" id="theme-template-label">
        {getLabel("theme.template")}
      </InputLabel>
      <Select
        data-testid="theme-template-select"
        label={getLabel("theme.template")}
        labelId="theme-template-label"
        name="themeTemplate"
        onChange={(event) => onChange(event.target.value as ThemeTemplateName)}
        value={value}
      >
        {availableTemplates.map((name) => {
          const label = getLabel(`theme.templates.${name}`);
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
};
