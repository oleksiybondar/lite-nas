import type {
  ServiceMetricContextValue,
  ServiceMetricUnitDTO,
} from "@dto/monitoring/service-metric";

/**
 * Contract required to render one service telemetry card.
 */
export type ServiceMetricCardProps = {
  /**
   * Service unit snapshot currently being displayed.
   */
  service: ServiceMetricUnitDTO;
  /**
   * Service action methods exposed by the polling provider.
   */
  serviceActions: Pick<
    ServiceMetricContextValue,
    "restart" | "toggleStartupBehaviour" | "toggleState"
  >;
};

export type ServiceMetricColorToken = "default" | "error" | "info" | "success";

export type ServiceMetricTone = {
  color: ServiceMetricColorToken;
  label: string;
};

export type ServiceMetricMetadataItem = {
  label: string;
  value: string;
};

export type ServiceMetricPrimaryAction = {
  color: "error" | "success";
  label: string;
  onClick: () => void;
};

export type ServiceMetricStartupAction = {
  isEnabled: boolean;
  label: string;
  labelColor: "error.main" | "info.main";
  onToggle: () => void;
};
