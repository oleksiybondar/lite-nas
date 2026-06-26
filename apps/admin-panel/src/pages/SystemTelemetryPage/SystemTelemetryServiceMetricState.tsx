import { ServiceMetricCard } from "@components/monitoring/ServiceMetricCard";
import { PaginationControl } from "@components/pagination";
import { useServiceMetric } from "@hooks/useServiceMetric";
import Button from "@mui/material/Button";
import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import MenuItem from "@mui/material/MenuItem";
import Paper from "@mui/material/Paper";
import Select from "@mui/material/Select";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

const allFilterValue = "";

/**
 * First-draft services state rendered from the snapshot polling provider.
 */
export const SystemTelemetryServiceMetricState = (): ReactElement => {
  const state = useServiceMetric();
  const activeServices = state.services.filter(
    (service) => service.active_state === "active",
  ).length;

  return (
    <Paper data-testid="service-metric-state-card" sx={{ p: 3 }}>
      <Stack spacing={2}>
        {renderServiceMetricHeader(state, activeServices)}
        {renderServiceMetricControls(state)}
        <Typography data-testid="service-metric-total-services" variant="body2">
          {state.visibleServicesCount} of {state.totalServices} services match the current search
          and filters.
        </Typography>
        {renderServiceMetricRows(state)}
        <PaginationControl
          pagination={{
            page: state.page,
            setPage: state.setPage,
            totalCount: state.visibleServicesCount,
            totalPages: state.totalPages,
          }}
          summary={`${state.visibleServicesCount} matching services across ${Math.max(state.totalPages, 1)} pages`}
          testIdPrefix="service-metric"
        />
        {renderServiceMetricPaginationState(state)}
      </Stack>
    </Paper>
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
};

/**
 * Renders the summary header for the current services snapshot state.
 */
const renderServiceMetricHeader = (
  state: ReturnType<typeof useServiceMetric>,
  activeServices: number,
): ReactElement => {
  return (
    <Stack spacing={1.5}>
      <Typography data-testid="service-metric-state-title" variant="h2">
        Services snapshot
      </Typography>
      <Typography color="text.secondary" data-testid="service-metric-state-summary" variant="body2">
        {buildServiceMetricSummary({
          activeServices,
          errorMessage: state.error?.message ?? null,
          isError: state.isError,
          isFetching: state.isFetching,
          isLoading: state.isLoading,
          totalServices: state.totalServices,
          visibleServicesCount: state.visibleServicesCount,
        })}
      </Typography>
    </Stack>
  );
};

/**
 * Renders the shared search and filter controls for the services snapshot view.
 */
const renderServiceMetricControls = (state: ReturnType<typeof useServiceMetric>): ReactElement => {
  return (
    <Stack direction={{ md: "row", xs: "column" }} spacing={1.5}>
      <TextField
        data-testid="service-metric-search-control"
        label="Search services"
        name="serviceMetricSearch"
        onChange={(event) => {
          state.setSearch(event.target.value);
        }}
        size="small"
        sx={{ minWidth: 240 }}
        value={state.search}
      />
      <FormControl size="small" sx={{ minWidth: 180 }}>
        <InputLabel id="service-metric-active-state-filter-label">Active state</InputLabel>
        <Select
          data-testid="service-metric-active-state-filter"
          label="Active state"
          labelId="service-metric-active-state-filter-label"
          onChange={(event) => {
            state.setActiveStateFilter(String(event.target.value));
          }}
          value={state.activeStateFilter}
        >
          <MenuItem value={allFilterValue}>All</MenuItem>
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
        <InputLabel id="service-metric-enabled-state-filter-label">Enabled state</InputLabel>
        <Select
          data-testid="service-metric-enabled-state-filter"
          label="Enabled state"
          labelId="service-metric-enabled-state-filter-label"
          onChange={(event) => {
            state.setEnabledStateFilter(String(event.target.value));
          }}
          value={state.enabledStateFilter}
        >
          <MenuItem value={allFilterValue}>All</MenuItem>
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
        Clear filters
      </Button>
    </Stack>
  );
};

/**
 * Renders the current page of service rows or the current empty state.
 */
const renderServiceMetricRows = (
  state: ReturnType<typeof useServiceMetric>,
): ReactElement | ReactElement[] => {
  if (state.services.length === 0) {
    return (
      <Typography data-testid="service-metric-empty-state" variant="body2">
        No services match the current search and filters.
      </Typography>
    );
  }

  return state.services.map((service) => {
    return <ServiceMetricCard key={service.name} service={service} serviceActions={state} />;
  });
};

/**
 * Renders the caption that exposes the current pagination state to tests and users.
 */
const renderServiceMetricPaginationState = (
  state: ReturnType<typeof useServiceMetric>,
): ReactElement => {
  return (
    <Typography
      color="text.secondary"
      data-testid="service-metric-pagination-state"
      variant="caption"
    >
      Page {state.page} of {Math.max(state.totalPages, 1)}. Showing up to {state.pageSize} services
      per page. Previous page available: {String(state.hasPreviousPage)}. Next page available:{" "}
      {String(state.hasNextPage)}.
    </Typography>
  );
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
}: BuildServiceMetricSummaryOptions): string => {
  if (isLoading) {
    return "Loading the latest service snapshot from the gateway.";
  }

  if (isError) {
    return errorMessage ?? "Failed to load the latest service snapshot.";
  }

  const refreshState = isFetching ? " A refresh is currently in progress." : "";
  return `${activeServices} active services on the current page. ${visibleServicesCount} of ${totalServices} services match the current search and filters.${refreshState}`;
};
