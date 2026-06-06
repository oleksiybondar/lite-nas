import { useAcknowledgeAlert } from "@domain/alerts/hooks/useAcknowledgeAlert";
import { useAlertsList } from "@domain/alerts/hooks/useAlertsList";
import { useAlerts } from "@hooks/useAlerts";
import { AlertsProvider } from "@providers/AlertsProvider";
import type { UseMutationResult, UseQueryResult } from "@tanstack/react-query";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import type { ReactElement } from "react";

vi.mock("@domain/alerts/hooks/useAlertsList", () => ({
  useAlertsList: vi.fn(),
}));

vi.mock("@domain/alerts/hooks/useAcknowledgeAlert", () => ({
  useAcknowledgeAlert: vi.fn(),
}));

describe("AlertsProvider", () => {
  test("exposes derived query state for one alerts route slice", () => {
    mockAlertsList();
    mockAcknowledgeAlert();

    renderAlertsProvider();

    expect(screen.getByTestId("alerts-domain")).toHaveTextContent("system");
    expect(screen.getByTestId("alerts-category")).toHaveTextContent("unacknowledged");
    expect(screen.getByTestId("alerts-api-path")).toHaveTextContent(
      "/api/alerts/system/unacknowledged?page=1&size=20",
    );
    expect(screen.getByTestId("alerts-route-path")).toHaveTextContent(
      "/alerts/system/unacknowledged",
    );
    expect(screen.getByTestId("alerts-total-count")).toHaveTextContent("3");
    expect(screen.getByTestId("alerts-total-pages")).toHaveTextContent("2");
  });

  test("resets page to one when filters change", async () => {
    mockAlertsList();
    mockAcknowledgeAlert();

    renderAlertsProvider();
    fireEvent.click(screen.getByTestId("set-page-two"));
    fireEvent.click(screen.getByTestId("set-source-filter"));

    await waitFor(() => {
      expect(screen.getByTestId("alerts-page")).toHaveTextContent("1");
    });
  });

  test("acknowledgeMany stops on first failure and refetches only after success", async () => {
    const refetch = vi.fn().mockResolvedValue({});
    const mutateAsync = vi
      .fn()
      .mockResolvedValueOnce(undefined)
      .mockRejectedValueOnce(new Error("Failed to acknowledge system alert."));

    mockAlertsList({ refetch });
    mockAcknowledgeAlert({ mutateAsync });

    renderAlertsProvider();

    await act(async () => {
      fireEvent.click(screen.getByTestId("acknowledge-many"));
    });

    expect(mutateAsync).toHaveBeenCalledTimes(2);
    expect(mutateAsync).toHaveBeenNthCalledWith(1, { id: "1" });
    expect(mutateAsync).toHaveBeenNthCalledWith(2, { id: "2" });
    expect(refetch).toHaveBeenCalledTimes(1);
  });
});

/**
 * Test-only consumer that exposes shared alerts context through stable selectors.
 */
const AlertsContextProbe = (): ReactElement => {
  const context = useAlerts();

  return (
    <>
      <span data-testid="alerts-domain">{context.domain}</span>
      <span data-testid="alerts-category">{context.category}</span>
      <span data-testid="alerts-api-path">{context.apiPath}</span>
      <span data-testid="alerts-route-path">{context.routePath}</span>
      <span data-testid="alerts-page">{context.page}</span>
      <span data-testid="alerts-total-count">{context.totalCount}</span>
      <span data-testid="alerts-total-pages">{context.totalPages}</span>
      <button data-testid="set-page-two" onClick={() => context.setPage(2)} type="button" />
      <button
        data-testid="set-source-filter"
        onClick={() => context.setSourceFilter(["collector"])}
        type="button"
      />
      <button
        data-testid="acknowledge-many"
        onClick={() => {
          void context.acknowledgeMany(["1", "2", "3"]);
        }}
        type="button"
      />
    </>
  );
};

/**
 * Renders the alerts provider around the shared test probe.
 */
const renderAlertsProvider = (): void => {
  render(
    <AlertsProvider category="unacknowledged" domain="system">
      <AlertsContextProbe />
    </AlertsProvider>,
  );
};

/**
 * Mocks the shared alerts list query hook.
 */
const mockAlertsList = ({
  refetch = vi.fn().mockResolvedValue({}),
}: {
  refetch?: () => Promise<unknown>;
} = {}): void => {
  vi.mocked(useAlertsList).mockReturnValue({
    data: {
      items: [],
      metadata: {
        page: 1,
        size: 20,
        total_count: 3,
        total_pages: 2,
      },
    },
    error: null,
    isError: false,
    isFetching: false,
    isLoading: false,
    refetch,
  } as UseQueryResult<ReturnType<typeof buildQueryData>>);
};

/**
 * Mocks the shared acknowledge-alert mutation hook.
 */
const mockAcknowledgeAlert = ({
  mutateAsync = vi.fn().mockResolvedValue(undefined),
}: {
  mutateAsync?: (input: { id: string }) => Promise<void>;
} = {}): void => {
  vi.mocked(useAcknowledgeAlert).mockReturnValue({
    isPending: false,
    mutateAsync,
  } as UseMutationResult<void, Error, { id: string }>);
};

/**
 * Builds one minimal alerts list payload for typed query-hook mocks.
 */
const buildQueryData = () => {
  return {
    items: [],
    metadata: {
      page: 1,
      size: 20,
      total_count: 3,
      total_pages: 2,
    },
  };
};
