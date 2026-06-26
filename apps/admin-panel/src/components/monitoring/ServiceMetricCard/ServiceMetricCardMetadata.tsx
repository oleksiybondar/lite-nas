import Chip from "@mui/material/Chip";
import Stack from "@mui/material/Stack";
import type { ReactElement } from "react";
import type { ServiceMetricMetadataItem } from "./types";

type ServiceMetricCardMetadataProps = {
  /**
   * Compact metadata fields shown on the third row of the card.
   */
  metadata: ServiceMetricMetadataItem[];
  /**
   * Service name used to build stable chip keys.
   */
  serviceName: string;
};

/**
 * Metadata row containing compact chips for service process and unit details.
 */
export const ServiceMetricCardMetadata = ({
  metadata,
  serviceName,
}: ServiceMetricCardMetadataProps): ReactElement => {
  return (
    <Stack direction="row" flexWrap="wrap" gap={1} useFlexGap>
      {metadata.map((item) => (
        <Chip
          data-test-class="service-metric-meta-chip"
          data-test-name={item.label}
          key={`${serviceName}-${item.label}`}
          label={`${item.label}: ${item.value}`}
          size="small"
          variant="outlined"
        />
      ))}
    </Stack>
  );
};
