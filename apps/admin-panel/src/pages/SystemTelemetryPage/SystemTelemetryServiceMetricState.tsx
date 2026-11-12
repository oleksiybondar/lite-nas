import { ServiceMetricCard } from "@components/monitoring/ServiceMetricCard";
import { useServiceMetric } from "@hooks/useServiceMetric";
import { useLanguage } from "@lite-nas/language";
import Button from "@mui/material/Button";
import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { SystemTelemetrySnapshotScaffold } from "./SystemTelemetrySnapshotScaffold";

const allFilterValue = "";

/**
 * First-draft services state rendered from the snapshot polling provider.
 */
export const SystemTelemetryServiceMetricState = (): ReactElement => {
  const state = useServiceMetric();
  const { getLabel } = useLanguage();
  const activeServices = state.services.filter(
    (service) => service.active_state === "active",
  ).length;

  return (
    <SystemTelemetrySnapshotScaffold
      content={renderServiceMetricRows(state, getLabel)}
      controls={renderServiceMetricControls(state, getLabel)}
      hasNextPage={state.hasNextPage}
      hasPreviousPage={state.hasPreviousPage}
      itemLabel={getLabel("admin.telemetry.service.itemLabel")}
      pageSize={state.pageSize}
      pagination={{
        page: state.page,
        setPage: state.setPage,
        totalCount: state.visibleServicesCount,
        totalPages: state.totalPages,
      }}
      summary={buildServiceMetricSummary({
        activeServices,
        errorMessage: state.error?.message ?? null,
        isError: state.isError,
        isFetching: state.isFetching,
        isLoading: state.isLoading,
        totalServices: state.totalServices,
        visibleServicesCount: state.visibleServicesCount,
        getLabel,
      })}
      testIdPrefix="service-metric"
      title={getLabel("admin.telemetry.service.title")}
      totalSummary={getLabel("admin.telemetry.service.totalSummary")
        .replace("{visible}", String(state.visibleServicesCount))
        .replace("{total}", String(state.totalServices))}
    />
  );
};

type BuildServiceMetricSummaryOptions = {
  activeServices: number;
  errorMessage: string | null;
  isError: boolean;
  isFetching: boolean;
  isLoading: boolean;
  totalServices: number;
  visibleServicesCount: number;
  getLabel: (key: string) => string;
};

/**
 * Renders the shared search and filter controls for the services snapshot view.
 */
const renderServiceMetricControls = (
  state: ReturnType<typeof useServiceMetric>,
  getLabel: (key: string) => string,
): ReactElement => {
  return (
    <Stack direction={{ md: "row", xs: "column" }} spacing={1.5}>
      <TextField
        data-testid="service-metric-search-control"
        label={getLabel("admin.telemetry.service.search")}
        name="serviceMetricSearch"
        onChange={(event) => {
          state.setSearch(event.target.value);
        }}
        size="small"
        sx={{ minWidth: 240 }}
        value={state.search}
      />
      <FormControl size="small" sx={{ minWidth: 180 }}>
        <InputLabel id="service-metric-active-state-filter-label">
          {getLabel("admin.telemetry.service.activeState")}
        </InputLabel>
        <Select
          data-testid="service-metric-active-state-filter"
          label={getLabel("admin.telemetry.service.activeState")}
          labelId="service-metric-active-state-filter-label"
          onChange={(event) => {
            state.setActiveStateFilter(String(event.target.value));
          }}
          value={state.activeStateFilter}
        >
          <MenuItem value={allFilterValue}>{getLabel("admin.telemetry.service.all")}</MenuItem>
          {state.availableActiveStates.map((value) => {
            return (
              <MenuItem key={value} value={value}>
                {value}
              </MenuItem>
            );
          })}
        </Select>
      </FormControl>
      <FormControl size="small" sx={{ minWidth: 180 }}>
        <InputLabel id="service-metric-enabled-state-filter-label">
          {getLabel("admin.telemetry.service.enabledState")}
        </InputLabel>
        <Select
          data-testid="service-metric-enabled-state-filter"
          label={getLabel("admin.telemetry.service.enabledState")}
          labelId="service-metric-enabled-state-filter-label"
          onChange={(event) => {
            state.setEnabledStateFilter(String(event.target.value));
          }}
          value={state.enabledStateFilter}
        >
          <MenuItem value={allFilterValue}>{getLabel("admin.telemetry.service.all")}</MenuItem>
          {state.availableEnabledStates.map((value) => {
            return (
              <MenuItem key={value} value={value}>
                {value}
              </MenuItem>
            );
          })}
        </Select>
      </FormControl>
      <Button
        data-testid="service-metric-clear-filters-button"
        onClick={state.clearFilters}
        variant="text"
      >
        {getLabel("admin.telemetry.service.clear")}
      </Button>
    </Stack>
  );
};

/**
 * Renders the current page of service rows or the current empty state.
 */
const renderServiceMetricRows = (
  state: ReturnType<typeof useServiceMetric>,
  getLabel: (key: string) => string,
): ReactElement | ReactElement[] => {
  if (state.services.length === 0) {
    return (
      <Typography data-testid="service-metric-empty-state" variant="body2">
        {getLabel("admin.telemetry.service.empty")}
      </Typography>
    );
  }

  return state.services.map((service) => {
    return <ServiceMetricCard key={service.name} service={service} serviceActions={state} />;
  });
};

/**
 * Resolves the short summary copy shown above the first-draft services snapshot preview.
 */
const buildServiceMetricSummary = ({
  activeServices,
  errorMessage,
  isError,
  isFetching,
  isLoading,
  totalServices,
  visibleServicesCount,
  getLabel,
}: BuildServiceMetricSummaryOptions): string => {
  if (isLoading) {
    return getLabel("admin.telemetry.service.loading");
  }

  if (isError) {
    return errorMessage ?? getLabel("admin.telemetry.service.loadFailed");
  }

  const refreshState = isFetching ? getLabel("admin.telemetry.service.refreshing") : "";
  return getLabel("admin.telemetry.service.summary")
    .replace("{active}", String(activeServices))
    .replace("{visible}", String(visibleServicesCount))
    .replace("{total}", String(totalServices))
    .replace("{refreshState}", refreshState);
};
