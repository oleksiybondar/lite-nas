import { AuthContext } from "@contexts/auth-context";
import type { ApiContextValue } from "@dto/api/api";
import type { AuthContextValue, AuthMeDTO, LoginRequestDTO } from "@dto/auth/auth";
import { refreshAuth } from "@helpers/auth-refresh";
import { createFetchRequestBuilder } from "@helpers/fetch-request-builder";
import { useApi } from "@hooks/useApi";
import {
  authFlowSchema,
  type PasswordConversation,
  passwordReplySchema,
} from "@lite-nas/authentication";
import { authDestination } from "@lite-nas/profile";
import { loginRequestSchema } from "@schemas/auth/login";
import { authMeResponseSchema } from "@schemas/auth/session";
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
const passwordStartPath = "/api/auth/password/start";
const passwordReplyPath = "/api/auth/password/reply";

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
  const [passwordConversation, setPasswordConversation] = useState<PasswordConversation | null>(
    null,
  );
  const loadMe = useLoadMe(get, applyMe, resetAuthState);
  const login = useLogin(loadMe, resetAuthState, setPasswordConversation);
  const logout = useLogout(resetAuthState);

  const replyPasswordChange = useReplyPasswordChange(
    passwordConversation,
    setPasswordConversation,
    loadMe,
  );
  const startPasswordChange = useStartPasswordChange();

  useEffect(() => {
    if (!isAuthInited) {
      void loadMe().catch(() => undefined);
    }
  }, [isAuthInited, loadMe]);

  useEffect(() => {
    window.addEventListener("lite-nas-admin-session-expired", resetAuthState);
    return () => window.removeEventListener("lite-nas-admin-session-expired", resetAuthState);
  }, [resetAuthState]);

  return {
    isAuthInited,
    isAuthenticated,
    login,
    logout,
    me,
    passwordConversation,
    replyPasswordChange,
    startPasswordChange,
  };
};

/** Sends PAM answers and updates the pending login conversation. */
const useReplyPasswordChange = (
  passwordConversation: PasswordConversation | null,
  setPasswordConversation: (conversation: PasswordConversation | null) => void,
  loadMe: () => Promise<Response>,
): NonNullable<AuthContextValue["replyPasswordChange"]> => {
  return useCallback(
    async (conversation, answer): Promise<Response> => {
      const payload = passwordReplySchema.parse({
        conversation_id: conversation.id,
        sequence: conversation.sequence,
        answer,
      });
      const response = await fetch(passwordReplyPath, {
        method: "POST",
        credentials: "include",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });
      return applyPasswordReply({
        response,
        conversation,
        currentConversation: passwordConversation,
        setPasswordConversation,
        loadMe,
      });
    },
    [passwordConversation, setPasswordConversation, loadMe],
  );
};

/** Applies a validated PAM reply and reloads the session after completed login. */
const applyPasswordReply = async ({
  response,
  conversation,
  currentConversation,
  setPasswordConversation,
  loadMe,
}: {
  response: Response;
  conversation: PasswordConversation;
  currentConversation: PasswordConversation | null;
  setPasswordConversation: (conversation: PasswordConversation | null) => void;
  loadMe: () => Promise<Response>;
}): Promise<Response> => {
  if (!response.ok) return response;
  const { data } = authFlowSchema.parse(await response.clone().json());
  if (currentConversation?.id === conversation.id)
    setPasswordConversation(data.conversation ?? null);
  if (data.authenticated && !data.conversation) await loadMe();
  return response;
};

/** Starts an on-demand PAM change using the current browser session. */
const useStartPasswordChange = (): NonNullable<AuthContextValue["startPasswordChange"]> => {
  return useCallback(async (): Promise<Response> => {
    const input: RequestInit = {
      method: "POST",
      credentials: "include",
      headers: { "Content-Type": "application/json" },
      body: "{}",
    };
    let response = await fetch(passwordStartPath, input);
    if (response.status === 401 && (await refreshAuth()).ok)
      response = await fetch(passwordStartPath, input);
    if (!response.ok) return response;
    authFlowSchema.parse(await response.clone().json());
    return response;
  }, []);
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

      const { data } = authMeResponseSchema.parse(await response.json());
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
  setPasswordConversation: (conversation: PasswordConversation | null) => void,
): AuthContextValue["login"] => {
  return useCallback(
    (loginValue: string, password: string, redirectTo?: string): Promise<Response> => {
      const payload = loginRequestSchema.parse({
        login: loginValue,
        password,
      }) satisfies LoginRequestDTO;

      const path =
        redirectTo === undefined
          ? loginPath
          : `${loginPath}?${new URLSearchParams({ redirect_to: authDestination(redirectTo) })}`;
      return createFetchRequestBuilder(path)
        .method("POST")
        .credentials("include")
        .payload(payload)
        .onSuccess((response) =>
          handleLoginSuccess(response, loadMe, resetAuthState, setPasswordConversation),
        )
        .onError(({ response }) => handleAuthFailure(response, resetAuthState))
        .execute();
    },
    [loadMe, resetAuthState, setPasswordConversation],
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
  setPasswordConversation: (conversation: PasswordConversation | null) => void,
): Promise<Response> => {
  const { data } = authFlowSchema.parse(await response.clone().json());

  if (data.conversation) {
    setPasswordConversation(data.conversation);
    return response;
  }

  setPasswordConversation(null);

  if (!data.authenticated) {
    resetAuthState();
    return response;
  }

  const current = await loadMe();
  return current.ok ? response : current;
};

/**
 * Resets auth state and returns the response that triggered the reset.
 */
const handleAuthFailure = (response: Response, resetAuthState: () => void): Response => {
  resetAuthState();
  return response;
};
