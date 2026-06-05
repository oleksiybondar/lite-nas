/**
 * Alert domains exposed by the gateway alerts API.
 */
export type AlertDomain = "security" | "system";

/**
 * Count payload returned by alert-count endpoints.
 */
export type AlertCountDTO = {
  /**
   * Total matching alert count.
   */
  count: number;
};

/**
 * Response envelope returned by alert-count endpoints.
 */
export type AlertCountResponseDTO = {
  /**
   * Common response metadata.
   */
  code?: string;
  /**
   * Count payload for the requested alert filter.
   */
  data: AlertCountDTO;
  /**
   * Common response metadata.
   */
  message?: string;
  /**
   * Common response metadata.
   */
  request_id?: string;
  /**
   * Whether the request completed successfully.
   */
  success: boolean;
  /**
   * Response creation timestamp.
   */
  timestamp: string;
  /**
   * Common response metadata.
   */
  trace_id?: string;
};
