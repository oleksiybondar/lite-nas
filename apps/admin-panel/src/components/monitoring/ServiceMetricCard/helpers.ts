import type { Theme } from "@mui/material/styles";
import { alpha } from "@mui/material/styles";
import type {
  ServiceMetricCardProps,
  ServiceMetricColorToken,
  ServiceMetricMetadataItem,
  ServiceMetricPrimaryAction,
  ServiceMetricStartupAction,
  ServiceMetricTone,
} from "./types";

const successSubstates = new Set(["running"]);
const failedValues = new Set(["failed", "failure"]);
const exitedSubstates = new Set(["dead", "exited"]);

/**
 * Resolves the visible label and semantic color for the service status row.
 */
export const resolveServiceStatusTone = (activeState?: string | null): ServiceMetricTone => {
  const normalized = normalizeMetricValue(activeState);

  if (normalized === "active" || normalized === "running") {
    return { color: "success", label: normalized };
  }

  if (failedValues.has(normalized)) {
    return { color: "error", label: normalized };
  }

  return { color: "default", label: normalized };
};

/**
 * Resolves the visible label and semantic color for the service sub-status row.
 */
export const resolveServiceSubstatusTone = (
  subState?: string | null,
  result?: string | null,
): ServiceMetricTone => {
  const normalizedSubState = normalizeMetricValue(subState);
  const normalizedResult = normalizeMetricValue(result);

  if (successSubstates.has(normalizedSubState)) {
    return { color: "success", label: normalizedSubState };
  }

  if (hasServiceFailure(normalizedSubState, normalizedResult)) {
    return buildFailedTone(normalizedSubState, normalizedResult);
  }

  if (isCleanExit(normalizedSubState, normalizedResult)) {
    return { color: "info", label: `${normalizedSubState} (${normalizedResult})` };
  }

  return { color: "default", label: normalizedSubState };
};

/**
 * Resolves the LED color shown beside the service name.
 */
export const resolveServiceLedColor = (
  theme: Theme,
  activeState?: string | null,
  result?: string | null,
): string => {
  const normalizedActiveState = normalizeMetricValue(activeState);
  const normalizedResult = normalizeMetricValue(result);

  if (hasServiceFailure(normalizedActiveState, normalizedResult)) {
    return theme.palette.error.main;
  }

  if (normalizedActiveState === "active" || normalizedActiveState === "running") {
    return theme.palette.success.main;
  }

  return theme.palette.action.disabled;
};

/**
 * Builds the chip-like colors used for state badges while keeping default values neutral.
 */
export const buildServiceToneSx = (theme: Theme, color: ServiceMetricColorToken) => {
  if (color === "success") {
    return {
      backgroundColor: alpha(theme.palette.success.main, 0.12),
      color: theme.palette.success.main,
    };
  }

  if (color === "info") {
    return {
      backgroundColor: alpha(theme.palette.info.main, 0.12),
      color: theme.palette.info.main,
    };
  }

  if (color === "error") {
    return {
      backgroundColor: alpha(theme.palette.error.main, 0.12),
      color: theme.palette.error.main,
    };
  }

  return {
    backgroundColor: theme.palette.action.hover,
    color: theme.palette.text.secondary,
  };
};

/**
 * Builds the compact metadata chips rendered on the third row of the service card.
 */
export const buildServiceMetadata = (
  service: ServiceMetricCardProps["service"],
): ServiceMetricMetadataItem[] => {
  return [
    { label: "Startup", value: formatServiceMetaValue(service.enabled_state) },
    { label: "Load", value: formatServiceMetaValue(service.load_state) },
    { label: "Manager", value: formatServiceMetaValue(service.manager) },
    { label: "Type", value: formatServiceMetaValue(service.unit_type) },
    { label: "Main PID", value: formatServiceMetaValue(service.main_pid) },
    { label: "Control PID", value: formatServiceMetaValue(service.control_pid) },
    { label: "PIDs", value: formatServicePids(service.pids) },
    { label: "CGroup", value: formatServiceMetaValue(service.cgroup) },
    { label: "Fragment", value: formatServiceMetaValue(service.fragment_path) },
  ].filter((item): item is ServiceMetricMetadataItem => item.value !== null);
};

/**
 * Resolves the primary start/stop action for the current service state.
 */
export const resolvePrimaryAction = (
  service: ServiceMetricCardProps["service"],
  serviceActions: ServiceMetricCardProps["serviceActions"],
): ServiceMetricPrimaryAction => {
  if (service.active_state === "active") {
    return {
      color: "error",
      label: "Stop",
      onClick: () => {
        void serviceActions.toggleState(service.name, "stop");
      },
    };
  }

  return {
    color: "success",
    label: "Start",
    onClick: () => {
      void serviceActions.toggleState(service.name, "start");
    },
  };
};

/**
 * Resolves the startup toggle contract for the current service.
 */
export const resolveStartupAction = (
  service: ServiceMetricCardProps["service"],
  serviceActions: ServiceMetricCardProps["serviceActions"],
): ServiceMetricStartupAction => {
  const isEnabled = service.enabled_state === "enabled";

  return {
    isEnabled,
    label: isEnabled ? "Disable" : "Enable",
    labelColor: isEnabled ? "error.main" : "info.main",
    onToggle: () => {
      void serviceActions.toggleStartupBehaviour(service.name, isEnabled ? "disable" : "enable");
    },
  };
};

/**
 * Formats one nullable metadata value for display in the compact service card rows.
 */
const formatServiceMetaValue = (value: number | string | null | undefined): string | null => {
  if (typeof value === "number") {
    return String(value);
  }

  if (typeof value === "string" && value.trim().length > 0) {
    return value;
  }

  return null;
};

/**
 * Formats service pid lists without rendering an empty marker.
 */
const formatServicePids = (pids?: number[] | null): string | null => {
  if (!Array.isArray(pids) || pids.length === 0) {
    return null;
  }

  return pids.join(", ");
};

/**
 * Normalizes service-state values into a stable lowercase label.
 */
const normalizeMetricValue = (value?: string | null): string => {
  if (typeof value !== "string" || value.trim().length === 0) {
    return "unknown";
  }

  return value.trim().toLowerCase();
};

const buildFailedTone = (subState: string, result: string): ServiceMetricTone => {
  return {
    color: "error",
    label: result !== "unknown" ? result : subState,
  };
};

const hasServiceFailure = (state: string, result: string): boolean => {
  return failedValues.has(state) || failedValues.has(result);
};

const isCleanExit = (subState: string, result: string): boolean => {
  return exitedSubstates.has(subState) && result === "success";
};
