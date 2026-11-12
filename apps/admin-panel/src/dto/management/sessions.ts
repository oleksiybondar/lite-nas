import type { HostSession } from "@schemas/management/sessions";

/** Actions available for a host session from the security inventory. */
export type SessionAction = "lock" | "unlock" | "terminate";

/** Session selected for a confirmed management action. */
export type SessionConfirmation = {
  session: HostSession;
  action: SessionAction;
};

/** State and operations exposed by the sessions page provider. */
export type SessionsContextValue = {
  allowed: boolean;
  actionError: string;
  confirmation: SessionConfirmation | null;
  isLoading: boolean;
  isError: boolean;
  page: number;
  pageCount: number;
  pending: string;
  sessions: HostSession[];
  totalSessions: number;
  visibleSessions: HostSession[];
  search: string;
  closeConfirmation: () => void;
  requestAction: (session: HostSession, action: SessionAction) => void;
  runAction: (session: HostSession, action: SessionAction) => Promise<void>;
  setPage: (page: number) => void;
  setSearch: (search: string) => void;
};

/** Actions carried beside the polled session inventory snapshot. */
export type SessionsPollingActions = {
  runSessionAction: (session: HostSession, action: SessionAction) => Promise<void>;
  actionError: string;
};
