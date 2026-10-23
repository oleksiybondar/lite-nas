import { AuthContext } from "@contexts/auth-context";
import { ApiProvider } from "@providers/ApiProvider";
import type { PropsWithChildren, ReactElement } from "react";

/** Supplies unauthenticated browser contexts for local preference component tests. */
export const TestProfileBoundary = ({ children }: PropsWithChildren): ReactElement => {
  return (
    <ApiProvider>
      <AuthContext.Provider
        value={{
          isAuthInited: true,
          isAuthenticated: false,
          me: null,
          login: vi.fn(),
          logout: vi.fn(),
        }}
      >
        {children}
      </AuthContext.Provider>
    </ApiProvider>
  );
};
