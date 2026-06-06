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
    <Stack
      alignItems={{ md: "flex-end", xs: "stretch" }}
      data-testid="alerts-filters-control"
      direction="row"
      flexWrap="wrap"
      gap={2}
      useFlexGap
    >
      {renderCategoryFilter({
        availableCategoryOptions,
        categoryFilter,
        setCategoryFilter,
      })}
      {renderSourceFilter({
        availableSourceOptions,
        setSourceFilter,
        sourceFilter,
      })}
      {renderPriorityFilter({
        availablePriorityOptions,
        priorityFilter,
        setPriorityFilter,
      })}
      {renderSeverityFilter({
        availableSeverityOptions,
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
 * Inputs required to render the category autocomplete filter.
 */
type CategoryFilterOptions = {
  availableCategoryOptions: AlertsControlPanelOption<string>[];
  categoryFilter: string[];
  setCategoryFilter: (value: string[]) => void;
};

/**
 * Builds the category filter with custom-entry autocomplete support.
 */
const renderCategoryFilter = ({
  availableCategoryOptions,
  categoryFilter,
  setCategoryFilter,
}: CategoryFilterOptions): ReactElement => {
  return (
    <AlertsControlPanelAutocompleteFilter
      label="Category"
      name="alertsCategoryFilter"
      onChange={setCategoryFilter}
      options={availableCategoryOptions}
      value={categoryFilter}
    />
  );
};

/**
 * Inputs required to render the source autocomplete filter.
 */
type SourceFilterOptions = {
  availableSourceOptions: AlertsControlPanelOption<string>[];
  setSourceFilter: (value: string[]) => void;
  sourceFilter: string[];
};

/**
 * Builds the source filter with custom-entry autocomplete support.
 */
const renderSourceFilter = ({
  availableSourceOptions,
  setSourceFilter,
  sourceFilter,
}: SourceFilterOptions): ReactElement => {
  return (
    <AlertsControlPanelAutocompleteFilter
      label="Source"
      name="alertsSourceFilter"
      onChange={setSourceFilter}
      options={availableSourceOptions}
      value={sourceFilter}
    />
  );
};

/**
 * Inputs required to render the fixed priority multiselect filter.
 */
type PriorityFilterOptions = {
  availablePriorityOptions: AlertsControlPanelOption<number>[];
  priorityFilter: number[];
  setPriorityFilter: (value: number[]) => void;
};

/**
 * Builds the priority filter backed by the fixed priority option set.
 */
const renderPriorityFilter = ({
  availablePriorityOptions,
  priorityFilter,
  setPriorityFilter,
}: PriorityFilterOptions): ReactElement => {
  return (
    <AlertsControlPanelMultiValueFilter
      label="Priority"
      name="alertsPriorityFilter"
      onChange={(value) => {
        setPriorityFilter(value.map(Number));
      }}
      options={availablePriorityOptions}
      value={priorityFilter.map(String)}
    />
  );
};

/**
 * Inputs required to render the fixed severity multiselect filter.
 */
type SeverityFilterOptions = {
  availableSeverityOptions: AlertsControlPanelOption<AlertSeverity>[];
  setSeverityFilter: (value: AlertSeverity[]) => void;
  severityFilter: AlertSeverity[];
};

/**
 * Builds the severity filter backed by the fixed severity option set.
 */
const renderSeverityFilter = ({
  availableSeverityOptions,
  setSeverityFilter,
  severityFilter,
}: SeverityFilterOptions): ReactElement => {
  return (
    <AlertsControlPanelMultiValueFilter
      label="Severity"
      name="alertsSeverityFilter"
      onChange={(value) => {
        setSeverityFilter(value as AlertSeverity[]);
      }}
      options={availableSeverityOptions}
      value={severityFilter}
    />
  );
};
