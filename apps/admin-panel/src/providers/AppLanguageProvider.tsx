import { useApi } from "@hooks/useApi";
import { useAuth } from "@hooks/useAuth";
import { LanguageProvider, type LanguageTransport } from "@lite-nas/language";
import { type PropsWithChildren, type ReactElement, useMemo } from "react";

/** Connects shared language preferences to the admin session and API pipeline. */
export const AppLanguageProvider = ({ children }: PropsWithChildren): ReactElement => {
  const { isAuthenticated, me } = useAuth();
  const { get, put } = useApi();
  const transport = useMemo<LanguageTransport>(
    () => ({
      read: (url) => get(url).execute(),
      write: (url, body) => put(url, body).execute(),
    }),
    [get, put],
  );
  return (
    <LanguageProvider userId={isAuthenticated ? (me?.user.id ?? null) : null} transport={transport}>
      {children}
    </LanguageProvider>
  );
};
