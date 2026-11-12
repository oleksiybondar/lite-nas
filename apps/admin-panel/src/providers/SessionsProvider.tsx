import { SessionsContext } from "@contexts/sessions-context";
import type {
  SessionAction,
  SessionConfirmation,
  SessionsContextValue,
  SessionsPollingActions,
} from "@dto/management/sessions";
import { useApi } from "@hooks/useApi";
import { usePollingResource } from "@hooks/usePollingResource";
import { useRbac } from "@hooks/useRbac";
import { useLanguage } from "@lite-nas/language";
import { PollingResourceProvider } from "@providers/PollingResourceProvider";
import { managementActionResponseSchema } from "@schemas/management/management-action";
import { type HostSession, sessionsResponseSchema } from "@schemas/management/sessions";
import { useQueryClient } from "@tanstack/react-query";
import type { PropsWithChildren, ReactElement } from "react";
import { useMemo, useState } from "react";

const sessionsQueryKey = ["polling-resource", "sessions", "snapshot"];

/** Parses the validated gateway response into the current session snapshot. */
const parseSessionsSnapshot = (value: unknown): HostSession[] =>
  sessionsResponseSchema.parse(value).data;

/** Provides the host session snapshot through the shared polling resource. */
export const SessionsProvider = ({ children }: PropsWithChildren): ReactElement => {
  const { getLabel } = useLanguage();
  const { post } = useApi();
  const { requireSecurity } = useRbac();
  const queryClient = useQueryClient();
  const [actionError, setActionError] = useState("");
  const allowed = requireSecurity();
  const actions = useMemo<SessionsPollingActions>(() => {
    // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The provider keeps session actions and their state updates together.
    const runSessionAction = async (session: HostSession, action: SessionAction): Promise<void> => {
      setActionError("");
      try {
        const response = await post(
          `/api/sessions/${encodeURIComponent(session.id)}/${action}`,
        ).execute();
        if (!response.ok) throw new Error(getLabel("admin.sessions.actionFailed"));
        managementActionResponseSchema.parse(await response.json());
        await queryClient.invalidateQueries({ queryKey: sessionsQueryKey });
      } catch (caught) {
        setActionError(getLabel("admin.sessions.actionFailed"));
        throw caught;
      }
    };

    return { actionError, runSessionAction };
  }, [actionError, getLabel, post, queryClient]);

  return (
    <PollingResourceProvider
      actions={actions}
      enabled={allowed}
      errorMessage={getLabel("admin.sessions.loadFailed")}
      parseResponse={parseSessionsSnapshot}
      path="/api/sessions"
      queryKey={sessionsQueryKey}
    >
      <SessionsStateProvider allowed={allowed}>{children}</SessionsStateProvider>
    </PollingResourceProvider>
  );
};

/** Adds session search, local pagination, and confirmation state to the shared snapshot. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The provider keeps session actions and their state updates together.
const SessionsStateProvider = ({
  allowed,
  children,
}: PropsWithChildren<{ allowed: boolean }>): ReactElement => {
  const base = usePollingResource<HostSession[], SessionsPollingActions>();
  const [search, setSearchValue] = useState("");
  const [page, setPage] = useState(1);
  const [pending, setPending] = useState("");
  const [confirmation, setConfirmation] = useState<SessionConfirmation | null>(null);
  const allSessions = base.snapshot ?? [];
  const normalizedSearch = search.toLocaleLowerCase();
  const sessions = allSessions.filter((session) =>
    [
      session.id,
      session.uid,
      session.user,
      session.seat,
      session.class,
      session.tty,
      session.leader,
    ]
      .join(" ")
      .toLocaleLowerCase()
      .includes(normalizedSearch),
  );
  const pageCount = Math.max(1, Math.ceil(sessions.length / 20));
  const currentPage = Math.min(page, pageCount);
  const visibleSessions = sessions.slice((currentPage - 1) * 20, currentPage * 20);

  const runAction = async (session: HostSession, action: SessionAction): Promise<void> => {
    setPending(session.id);
    try {
      await base.runSessionAction(session, action);
      setConfirmation(null);
    } finally {
      setPending("");
    }
  };

  const value: SessionsContextValue = {
    allowed,
    actionError: base.actionError,
    confirmation,
    isLoading: base.isLoading,
    isError: base.isError,
    page: currentPage,
    pageCount,
    pending,
    sessions,
    totalSessions: allSessions.length,
    visibleSessions,
    search,
    closeConfirmation: () => setConfirmation(null),
    requestAction: (session, action) => setConfirmation({ session, action }),
    runAction,
    setPage,
    setSearch: (value) => {
      setSearchValue(value);
      setPage(1);
    },
  };

  return <SessionsContext.Provider value={value}>{children}</SessionsContext.Provider>;
};
