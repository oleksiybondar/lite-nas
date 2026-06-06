import { AlertsControlPanelAutocompleteFilter } from "@components/alerts/AlertsControlPanel/AlertsControlPanelAutocompleteFilter";
import { AlertsControlPanelMultiValueFilter } from "@components/alerts/AlertsControlPanel/AlertsControlPanelMultiValueFilter";
import type { AlertSeverity, AlertsControlPanelOption } from "@dto/alerts/alerts";
import { useAlertsControlPanel } from "@hooks/useAlertsControlPanel";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";

/**
 * Renders alerts page filters for category, source, priority, and severity.
 */
export const AlertsControlPanelFilters = (): ReactElement => {
  const {
    availableCategoryOptions,
    availablePriorityOptions,
    availableSeverityOptions,
    availableSourceOptions,
    categoryFilter,
    clearFilters,
    priorityFilter,
    setCategoryFilter,
    setPriorityFilter,
    setSeverityFilter,
    setSourceFilter,
    severityFilter,
    sourceFilter,
  } = useAlertsControlPanel();

  return (
    <Stack data-testid="alerts-filters-control" spacing={2}>
      {renderStringFilterRow({
        availableCategoryOptions,
        availableSourceOptions,
        categoryFilter,
        setCategoryFilter,
        setSourceFilter,
        sourceFilter,
      })}
      {renderPrioritySeverityFilterRow({
        availablePriorityOptions,
        availableSeverityOptions,
        priorityFilter,
        setPriorityFilter,
        setSeverityFilter,
        severityFilter,
      })}
      <Stack alignItems="flex-end" direction="row" justifyContent="flex-end">
        <Button data-testid="alerts-clear-filters-button" onClick={clearFilters} variant="text">
          Clear filters
        </Button>
      </Stack>
    </Stack>
  );
};

/**
 * Inputs required to render the category/source autocomplete filter row.
 */
type StringFilterRowOptions = {
  availableCategoryOptions: AlertsControlPanelOption<string>[];
  availableSourceOptions: AlertsControlPanelOption<string>[];
  categoryFilter: string[];
  setCategoryFilter: (value: string[]) => void;
  setSourceFilter: (value: string[]) => void;
  sourceFilter: string[];
};

/**
 * Builds the category/source filter row with custom-entry autocomplete inputs.
 */
const renderStringFilterRow = ({
  availableCategoryOptions,
  availableSourceOptions,
  categoryFilter,
  setCategoryFilter,
  setSourceFilter,
  sourceFilter,
}: StringFilterRowOptions): ReactElement => {
  return (
    <Stack
      alignItems={{ md: "center", xs: "stretch" }}
      direction={{ md: "row", xs: "column" }}
      spacing={2}
    >
      <AlertsControlPanelAutocompleteFilter
        label="Category"
        name="alertsCategoryFilter"
        onChange={setCategoryFilter}
        options={availableCategoryOptions}
        value={categoryFilter}
      />
      <AlertsControlPanelAutocompleteFilter
        label="Source"
        name="alertsSourceFilter"
        onChange={setSourceFilter}
        options={availableSourceOptions}
        value={sourceFilter}
      />
    </Stack>
  );
};

/**
 * Inputs required to render the fixed priority/severity multiselect row.
 */
type PrioritySeverityFilterRowOptions = {
  availablePriorityOptions: AlertsControlPanelOption<number>[];
  availableSeverityOptions: AlertsControlPanelOption<AlertSeverity>[];
  priorityFilter: number[];
  setPriorityFilter: (value: number[]) => void;
  setSeverityFilter: (value: AlertSeverity[]) => void;
  severityFilter: AlertSeverity[];
};

/**
 * Builds the priority/severity filter row backed by fixed option sets.
 */
const renderPrioritySeverityFilterRow = ({
  availablePriorityOptions,
  availableSeverityOptions,
  priorityFilter,
  setPriorityFilter,
  setSeverityFilter,
  severityFilter,
}: PrioritySeverityFilterRowOptions): ReactElement => {
  return (
    <Stack
      alignItems={{ md: "center", xs: "stretch" }}
      direction={{ md: "row", xs: "column" }}
      spacing={2}
    >
      <AlertsControlPanelMultiValueFilter
        label="Priority"
        name="alertsPriorityFilter"
        onChange={(value) => {
          setPriorityFilter(value.map(Number));
        }}
        options={availablePriorityOptions}
        value={priorityFilter.map(String)}
      />
      <AlertsControlPanelMultiValueFilter
        label="Severity"
        name="alertsSeverityFilter"
        onChange={(value) => {
          setSeverityFilter(value as AlertSeverity[]);
        }}
        options={availableSeverityOptions}
        value={severityFilter}
      />
    </Stack>
  );
};
