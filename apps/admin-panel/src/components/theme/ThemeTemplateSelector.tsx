import { useThemeManager } from "@hooks/useThemeManager";

import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import type { ThemeSource, ThemeTemplateName } from "@theme/index";
import type { ReactElement } from "react";

const formatThemeTemplateLabel = (templateName: ThemeTemplateName): string => {
  return templateName.slice(0, 1).toUpperCase() + templateName.slice(1);
};

interface ThemeTemplateSelectorProps {
  disabled?: boolean;
  onChange?: (templateName: ThemeTemplateName) => void;
  source?: ThemeSource;
  value?: ThemeTemplateName;
}

export const ThemeTemplateSelector = ({
  disabled,
  onChange,
  source: sourceOverride,
  value,
}: ThemeTemplateSelectorProps): ReactElement => {
  const { availableTemplates, setTemplateName, source, templateName } = useThemeManager();
  const selectedSource = sourceOverride ?? source;
  const selectedValue = value ?? templateName;
  const isDisabled = disabled ?? selectedSource !== "user";

  return (
    <FormControl data-testid="theme-template-control" disabled={isDisabled} fullWidth>
      <InputLabel data-testid="theme-template-label" id="theme-template-label">
        Theme template
      </InputLabel>
      <Select
        data-testid="theme-template-select"
        label="Theme template"
        labelId="theme-template-label"
        name="themeTemplate"
        onChange={(event) => {
          const nextValue = event.target.value as ThemeTemplateName;

          if (onChange !== undefined) {
            onChange(nextValue);
            return;
          }

          setTemplateName(nextValue);
        }}
        value={selectedValue}
      >
        {availableTemplates.map((availableTemplate) => {
          return (
            <MenuItem
              data-test-class="theme-template-option"
              data-test-name={formatThemeTemplateLabel(availableTemplate)}
              key={availableTemplate}
              value={availableTemplate}
            >
              {formatThemeTemplateLabel(availableTemplate)}
            </MenuItem>
          );
        })}
      </Select>
    </FormControl>
  );
};
