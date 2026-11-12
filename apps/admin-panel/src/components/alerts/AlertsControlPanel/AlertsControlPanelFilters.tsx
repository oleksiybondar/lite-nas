import { AlertsControlPanelAutocompleteFilter } from "@components/alerts/AlertsControlPanel/AlertsControlPanelAutocompleteFilter";
import { AlertsControlPanelMultiValueFilter } from "@components/alerts/AlertsControlPanel/AlertsControlPanelMultiValueFilter";
import type { AlertSeverity, AlertsControlPanelOption } from "@dto/alerts/alerts";
import { useAlertsControlPanel } from "@hooks/useAlertsControlPanel";
import { useLanguage } from "@lite-nas/language";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";

/**
 * Renders alerts page filters for category, source, priority, and severity.
 */
export const AlertsControlPanelFilters = (): ReactElement => {
  const { getLabel } = useLanguage();
  const filters = useAlertsControlPanel();

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
        availableCategoryOptions: filters.availableCategoryOptions,
        categoryFilter: filters.categoryFilter,
        setCategoryFilter: filters.setCategoryFilter,
        getLabel,
      })}
      {renderSourceFilter({
        availableSourceOptions: filters.availableSourceOptions,
        setSourceFilter: filters.setSourceFilter,
        sourceFilter: filters.sourceFilter,
        getLabel,
      })}
      {renderPriorityFilter({
        availablePriorityOptions: filters.availablePriorityOptions,
        priorityFilter: filters.priorityFilter,
        setPriorityFilter: filters.setPriorityFilter,
        getLabel,
      })}
      {renderSeverityFilter({
        availableSeverityOptions: filters.availableSeverityOptions,
        setSeverityFilter: filters.setSeverityFilter,
        severityFilter: filters.severityFilter,
        getLabel,
      })}
      {renderClearFiltersButton(filters.clearFilters, getLabel)}
    </Stack>
  );
};

/** Renders the action that resets every alert filter. */
const renderClearFiltersButton = (
  clearFilters: () => void,
  getLabel: (key: string) => string,
): ReactElement => (
  <Stack alignItems="flex-end" direction="row" justifyContent="flex-end">
    <Button data-testid="alerts-clear-filters-button" onClick={clearFilters} variant="text">
      {getLabel("admin.alerts.clearFilters")}
    </Button>
  </Stack>
);

/**
 * Inputs required to render the category autocomplete filter.
 */
type CategoryFilterOptions = {
  availableCategoryOptions: AlertsControlPanelOption<string>[];
  categoryFilter: string[];
  setCategoryFilter: (value: string[]) => void;
  getLabel: (key: string) => string;
};

/**
 * Builds the category filter with custom-entry autocomplete support.
 */
const renderCategoryFilter = ({
  availableCategoryOptions,
  categoryFilter,
  setCategoryFilter,
  getLabel,
}: CategoryFilterOptions): ReactElement => {
  return (
    <AlertsControlPanelAutocompleteFilter
      label={getLabel("admin.alerts.category")}
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
  getLabel: (key: string) => string;
};

/**
 * Builds the source filter with custom-entry autocomplete support.
 */
const renderSourceFilter = ({
  availableSourceOptions,
  setSourceFilter,
  sourceFilter,
  getLabel,
}: SourceFilterOptions): ReactElement => {
  return (
    <AlertsControlPanelAutocompleteFilter
      label={getLabel("admin.alerts.source")}
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
  getLabel: (key: string) => string;
};

/**
 * Builds the priority filter backed by the fixed priority option set.
 */
const renderPriorityFilter = ({
  availablePriorityOptions,
  priorityFilter,
  setPriorityFilter,
  getLabel,
}: PriorityFilterOptions): ReactElement => {
  return (
    <AlertsControlPanelMultiValueFilter
      label={getLabel("admin.alerts.priority")}
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
  getLabel: (key: string) => string;
};

/**
 * Builds the severity filter backed by the fixed severity option set.
 */
const renderSeverityFilter = ({
  availableSeverityOptions,
  setSeverityFilter,
  severityFilter,
  getLabel,
}: SeverityFilterOptions): ReactElement => {
  return (
    <AlertsControlPanelMultiValueFilter
      label={getLabel("admin.alerts.severity")}
      formatSelection={(values) =>
        values
          .map((value) => getLabel(`admin.alertTable.severity.${value as AlertSeverity}`))
          .join(", ")
      }
      name="alertsSeverityFilter"
      onChange={(value) => {
        setSeverityFilter(value as AlertSeverity[]);
      }}
      options={availableSeverityOptions.map((option) => ({
        ...option,
        label: getLabel(`admin.alertTable.severity.${option.value}`),
      }))}
      value={severityFilter}
    />
  );
};
