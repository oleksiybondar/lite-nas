import { PublicAppLayout } from "@components/layout/PublicAppLayout";
import { useAuth } from "@hooks/useAuth";
import { AuthenticationForm, authFlowSchema } from "@lite-nas/authentication";
import { useLanguage } from "@lite-nas/language";
import { authDestination, loginDestination } from "@lite-nas/profile";
import Container from "@mui/material/Container";
import Paper from "@mui/material/Paper";
import { authSessionResponseSchema } from "@schemas/auth/session";
import type { FormEvent, ReactElement } from "react";
import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";

/**
 * Browser login page rendered when the auth guard resolves an anonymous state.
 */
export const LoginPage = (): ReactElement => {
  const { getLabel } = useLanguage();
  const { login, passwordConversation, replyPasswordChange } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const from = (location.state as { from?: unknown } | null)?.from;
  const destination = loginDestination(location.search, from);
  const { error, isSubmitting, onSubmit, onAnswer, password, setPassword, setUsername, username } =
    useLoginForm({
      login,
      replyPasswordChange,
      passwordConversation: passwordConversation ?? null,
      destination,
      onAuthenticated: (returned) => {
        void navigate(authDestination(returned ?? destination), { replace: true });
      },
    });

  return (
    <PublicAppLayout>
      <Container data-testid="login-page" maxWidth="xs" sx={{ py: 8 }}>
        <Paper data-testid="login-card" sx={{ p: 4 }}>
          <AuthenticationForm
            title={getLabel("admin.signIn")}
            username={username}
            password={password}
            setUsername={setUsername}
            setPassword={setPassword}
            onSubmit={onSubmit}
            onAnswer={onAnswer}
            conversation={passwordConversation ?? null}
            busy={isSubmitting}
            error={error}
            testIds={{
              form: "login-form",
              title: "login-title",
              username: "login-username-field",
              password: "login-password-field",
              submit: "login-submit-button",
              error: "login-error-message",
            }}
          />
        </Paper>
      </Container>
    </PublicAppLayout>
  );
};

/**
 * Login function supplied by the auth context.
 */
type LoginAction = ReturnType<typeof useAuth>["login"];

/**
 * Owns login form state and submit side effects.
 */
const useLoginForm = ({
  login,
  replyPasswordChange,
  passwordConversation,
  destination,
  onAuthenticated,
}: {
  login: LoginAction;
  replyPasswordChange: ReturnType<typeof useAuth>["replyPasswordChange"];
  passwordConversation: ReturnType<typeof useAuth>["passwordConversation"];
  destination: string;
  onAuthenticated: (redirectTo?: string) => void;
}) => {
  const { getLabel } = useLanguage();
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [password, setPassword] = useState("");
  const [username, setUsername] = useState("");
  const reply =
    replyPasswordChange ?? (() => Promise.reject(new Error("Password change unavailable.")));

  const onSubmit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      const response = await login(username, password, destination);
      await handleLoginFormResponse(
        response,
        onAuthenticated,
        setError,
        getLabel("admin.loginFailed"),
      );
    } catch {
      setError(getLabel("admin.loginFailed"));
    } finally {
      setIsSubmitting(false);
    }
  };

  const onAnswer = async (answer: string): Promise<void> => {
    if (!passwordConversation) return;
    setError(null);
    setIsSubmitting(true);
    try {
      const response = await reply(passwordConversation, answer);
      await handlePasswordAnswerResponse({
        response,
        onAuthenticated,
        setError,
        destination,
        failureMessage: getLabel("common.passwordChangeFailed"),
      });
    } catch {
      setError(getLabel("common.passwordChangeFailed"));
    } finally {
      setIsSubmitting(false);
    }
  };

  return { error, isSubmitting, onSubmit, onAnswer, password, setPassword, setUsername, username };
};

/** Applies a completed login response to the form. */
const handleLoginFormResponse = async (
  response: Response,
  onAuthenticated: (redirectTo?: string) => void,
  setError: (message: string) => void,
  failureMessage: string,
): Promise<void> => {
  if (!response.ok) {
    setError(failureMessage);
    return;
  }
  const { data } = authFlowSchema.parse(await response.clone().json());
  if (data.authenticated) onAuthenticated(await readLoginRedirect(response));
};

/** Applies the next PAM prompt or final password-change result to the form. */
const handlePasswordAnswerResponse = async ({
  response,
  onAuthenticated,
  setError,
  destination,
  failureMessage,
}: {
  response: Response;
  onAuthenticated: (redirectTo?: string) => void;
  setError: (message: string) => void;
  destination: string;
  failureMessage: string;
}): Promise<void> => {
  if (!response.ok) {
    setError(failureMessage);
    return;
  }
  const { data } = authFlowSchema.parse(await response.clone().json());
  if (data.status === "denied") {
    setError(passwordAnswerError(data.messages, failureMessage));
    return;
  }
  if (data.authenticated && !data.conversation) onAuthenticated(data.redirect_to ?? destination);
};

/** Chooses a PAM failure message while retaining localized fallback text. */
const passwordAnswerError = (messages: { text: string }[] | undefined, fallback: string): string =>
  messages?.map((message) => message.text).join(" ") || fallback;

/** Reads the echoed path while preserving login links across older gateway deployments. */
const readLoginRedirect = async (response: Response): Promise<string | undefined> => {
  try {
    return authSessionResponseSchema.parse(await response.clone().json()).data.redirect_to;
  } catch {
    return undefined;
  }
};
