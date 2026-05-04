/**
 * Browser-facing auth principal returned by `/api/auth/me`.
 */
export type AuthMeUserDTO = {
  /**
   * Stable user identifier.
   */
  id: string;
  /**
   * Login name displayed for the authenticated principal.
   */
  login: string;
};

/**
 * Current authenticated session state returned by `/api/auth/me`.
 */
export type AuthMeDTO = {
  /**
   * Whether the gateway resolved the request to an authenticated session.
   */
  authenticated: boolean;
  /**
   * Gateway auth mechanism used for the current session.
   */
  auth_type: string;
  /**
   * Role names assigned to the authenticated principal.
   */
  roles: string[];
  /**
   * Permission scopes assigned to the authenticated principal.
   */
  scopes: string[];
  /**
   * Authenticated principal identity.
   */
  user: AuthMeUserDTO;
};

/**
 * Response envelope returned by `/api/auth/me`.
 */
export type AuthMeResponseDTO = {
  /**
   * Current authenticated session state.
   */
  data: AuthMeDTO;
};

/**
 * Login request body submitted to `/api/auth/login`.
 */
export type LoginRequestDTO = {
  /**
   * Local login name or email address.
   */
  login: string;
  /**
   * Password input for the login flow.
   */
  password: string;
};

/**
 * User payload returned by session-issuing auth endpoints.
 */
export type AuthSessionUserDTO = {
  /**
   * Stable user identifier.
   */
  id: string;
};

/**
 * Session state returned by login and refresh endpoints.
 */
export type AuthSessionDTO = {
  /**
   * Whether the gateway issued an authenticated browser session.
   */
  authenticated: boolean;
  /**
   * Authenticated principal identity available in session responses.
   */
  user: AuthSessionUserDTO;
};

/**
 * Response envelope returned by `/api/auth/login` and `/api/auth/refresh`.
 */
export type AuthSessionResponseDTO = {
  /**
   * Issued browser session state.
   */
  data: AuthSessionDTO;
};

/**
 * Auth context state and commands exposed by `AuthProvider`.
 */
export type AuthContextValue = {
  /**
   * Whether the initial `/api/auth/me` check has completed for this SPA session.
   */
  isAuthInited: boolean;
  /**
   * Whether the current browser state is authenticated according to the gateway.
   */
  isAuthenticated: boolean;
  /**
   * Current authenticated principal state, or `null` when anonymous.
   */
  me: AuthMeDTO | null;
  /**
   * Attempts to create a browser session with the submitted credentials.
   */
  login: (login: string, password: string) => Promise<Response>;
  /**
   * Attempts to expire the current browser session.
   */
  logout: () => Promise<Response>;
};
