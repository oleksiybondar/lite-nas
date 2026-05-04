import { AuthContext } from "@contexts/auth-context";
import type { ApiContextValue } from "@dto/api/api";
import type {
  AuthContextValue,
  AuthMeDTO,
  AuthMeResponseDTO,
  AuthSessionResponseDTO,
  LoginRequestDTO,
} from "@dto/auth/auth";
import { createFetchRequestBuilder } from "@helpers/fetch-request-builder";
import { useApi } from "@hooks/useApi";
import type { PropsWithChildren, ReactElement } from "react";
import { useCallback, useEffect, useState } from "react";

/**
 * Current-session endpoint owned by the web gateway.
 */
const mePath = "/api/auth/me";

/**
 * Login endpoint that issues auth cookies.
 */
const loginPath = "/api/auth/login";

/**
 * Logout endpoint that expires auth cookies.
 */
const logoutPath = "/api/auth/logout";

/**
 * Provides browser auth state and auth commands to the admin panel.
 */
export const AuthProvider = ({ children }: PropsWithChildren): ReactElement => {
  const value = useAuthProviderValue();

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

/**
 * Builds the value exposed by `AuthContext`.
 */
const useAuthProviderValue = (): AuthContextValue => {
  const { get } = useApi();
  const { applyMe, isAuthInited, isAuthenticated, me, resetAuthState } = useAuthState();
  const loadMe = useLoadMe(get, applyMe, resetAuthState);
  const login = useLogin(loadMe, resetAuthState);
  const logout = useLogout(resetAuthState);

  useEffect(() => {
    if (!isAuthInited) {
      void loadMe();
    }
  }, [isAuthInited, loadMe]);

  return { isAuthInited, isAuthenticated, login, logout, me };
};

/**
 * Owns browser auth state and local auth-state transitions.
 */
const useAuthState = () => {
  const [isAuthInited, setIsAuthInited] = useState(false);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [me, setMe] = useState<AuthMeDTO | null>(null);

  const resetAuthState = useCallback((): void => {
    setMe(null);
    setIsAuthenticated(false);
    setIsAuthInited(true);
  }, []);

  const applyMe = useCallback((nextMe: AuthMeDTO): void => {
    setMe(nextMe.authenticated ? nextMe : null);
    setIsAuthenticated(nextMe.authenticated);
    setIsAuthInited(true);
  }, []);

  return { applyMe, isAuthInited, isAuthenticated, me, resetAuthState };
};

/**
 * Creates the current-user loader backed by the app API client.
 */
const useLoadMe = (
  get: ApiContextValue["get"],
  applyMe: (nextMe: AuthMeDTO) => void,
  resetAuthState: () => void,
): (() => Promise<Response>) => {
  return useCallback(async (): Promise<Response> => {
    try {
      const response = await get(mePath).execute();

      if (!response.ok) {
        resetAuthState();
        return response;
      }

      const { data } = (await response.json()) as AuthMeResponseDTO;
      applyMe(data);
      return response;
    } catch (error) {
      resetAuthState();
      throw error;
    }
  }, [applyMe, get, resetAuthState]);
};

/**
 * Creates the login command backed by a raw fetch request.
 */
const useLogin = (
  loadMe: () => Promise<Response>,
  resetAuthState: () => void,
): AuthContextValue["login"] => {
  return useCallback(
    (loginValue: string, password: string): Promise<Response> => {
      const payload: LoginRequestDTO = { login: loginValue, password };

      return createFetchRequestBuilder(loginPath)
        .method("POST")
        .credentials("include")
        .payload(payload)
        .onSuccess((response) => handleLoginSuccess(response, loadMe, resetAuthState))
        .onError(({ response }) => handleAuthFailure(response, resetAuthState))
        .execute();
    },
    [loadMe, resetAuthState],
  );
};

/**
 * Creates the logout command backed by a raw fetch request.
 */
const useLogout = (resetAuthState: () => void): AuthContextValue["logout"] => {
  return useCallback((): Promise<Response> => {
    return createFetchRequestBuilder(logoutPath)
      .method("POST")
      .credentials("include")
      .payload({})
      .onSuccess((response) => handleAuthFailure(response, resetAuthState))
      .onError(({ response }) => handleAuthFailure(response, resetAuthState))
      .execute();
  }, [resetAuthState]);
};

/**
 * Applies login response state and reloads the full current-user payload.
 */
const handleLoginSuccess = async (
  response: Response,
  loadMe: () => Promise<Response>,
  resetAuthState: () => void,
): Promise<Response> => {
  const { data } = (await response.clone().json()) as AuthSessionResponseDTO;

  if (!data.authenticated) {
    resetAuthState();
    return response;
  }

  await loadMe();
  return response;
};

/**
 * Resets auth state and returns the response that triggered the reset.
 */
const handleAuthFailure = (response: Response, resetAuthState: () => void): Response => {
  resetAuthState();
  return response;
};
