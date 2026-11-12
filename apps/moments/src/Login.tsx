import {
  AuthenticationForm,
  type AuthFlow,
  type PasswordConversation,
} from "@lite-nas/authentication";
import { useLanguage } from "@lite-nas/language";
import { loginDestination } from "@lite-nas/profile";
import Container from "@mui/material/Container";
import Paper from "@mui/material/Paper";
import type { FormEvent, ReactElement } from "react";
import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { beginLogin, loadMe, replyPasswordChange } from "./api";
import { MomentsLayout } from "./MomentsLayout";
import { MomentsPublicHeader } from "./MomentsPublicHeader";
import type { Me } from "./schemas";

/** Collects credentials and renders the current PAM authentication step. */
export const Login = ({ onSuccess }: { onSuccess: (me: Me) => void }): ReactElement => {
  const { getLabel } = useLanguage();
  const location = useLocation();
  const navigate = useNavigate();
  const form = useMomentsLoginForm({
    destination: loginDestination(location.search),
    onSuccess,
    navigate,
    signInError: getLabel("common.signInFailed"),
    passwordError: getLabel("common.passwordChangeFailed"),
  });
  return (
    <MomentsLayout header={<MomentsPublicHeader />}>
      <Container data-testid="moments-login" maxWidth="xs" sx={{ py: 8, my: "auto" }}>
        <Paper sx={{ p: { xs: 3, sm: 4 } }}>
          <AuthenticationForm
            title={getLabel("moments.signIn")}
            username={form.username}
            password={form.password}
            setUsername={form.setUsername}
            setPassword={form.setPassword}
            onSubmit={form.onSubmit}
            onAnswer={form.onAnswer}
            conversation={form.conversation}
            busy={form.isSubmitting}
            error={form.error}
            testIds={{
              username: "moments-login-name",
              password: "moments-login-password",
              submit: "moments-login-submit",
            }}
          />
        </Paper>
      </Container>
    </MomentsLayout>
  );
};

/** State and navigation needed to complete the browser login flow. */
type LoginFormOptions = {
  destination: string;
  onSuccess: (me: Me) => void;
  navigate: ReturnType<typeof useNavigate>;
  signInError: string;
  passwordError: string;
};

/** Form inputs and state updates shared by the two PAM requests. */
type LoginActionOptions = LoginFormOptions & {
  username: string;
  password: string;
  conversation: PasswordConversation | null;
  setPassword: (value: string) => void;
  setError: (value: string | null) => void;
  setIsSubmitting: (value: boolean) => void;
  setConversation: (value: PasswordConversation | null) => void;
};

/** Holds the form fields while its actions handle login and password prompts. */
const useMomentsLoginForm = (options: LoginFormOptions) => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [conversation, setConversation] = useState<PasswordConversation | null>(null);
  const { onSubmit, onAnswer } = useLoginActions({
    ...options,
    username,
    password,
    conversation,
    setPassword,
    setError,
    setIsSubmitting,
    setConversation,
  });
  return {
    username,
    setUsername,
    password,
    setPassword,
    error,
    isSubmitting,
    conversation,
    onSubmit,
    onAnswer,
  };
};

/** Runs either PAM request with the same loading and error handling. */
const useLoginActions = (options: LoginActionOptions) => {
  const run = async (
    request: () => Promise<AuthFlow>,
    apply: (result: AuthFlow, state: LoginActionOptions) => Promise<void>,
    fallback: string,
  ): Promise<void> => {
    options.setError(null);
    options.setIsSubmitting(true);
    try {
      await apply(await request(), options);
    } catch {
      options.setError(fallback);
    } finally {
      options.setIsSubmitting(false);
    }
  };
  const onSubmit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault();
    await run(
      () =>
        beginLogin({ login: options.username, password: options.password }, options.destination),
      applyLoginResult,
      options.signInError,
    );
  };
  const onAnswer = async (answer: string): Promise<void> => {
    const conversation = options.conversation;
    if (!conversation) return;
    await run(
      () => replyPasswordChange(conversation, answer),
      applyAnswerResult,
      options.passwordError,
    );
  };
  return { onSubmit, onAnswer };
};

/** Applies the initial login outcome or preserves the PAM change prompt. */
const applyLoginResult = async (result: AuthFlow, options: LoginActionOptions): Promise<void> => {
  if (result.conversation) {
    options.setConversation(result.conversation);
    options.setPassword("");
    return;
  }
  if (!result.authenticated) {
    options.setError(options.signInError);
    return;
  }
  await finishSignIn(options, result.redirect_to ?? options.destination);
};

/** Applies a password-change reply and displays PAM's failure message. */
const applyAnswerResult = async (result: AuthFlow, options: LoginActionOptions): Promise<void> => {
  if (result.conversation) {
    options.setConversation(result.conversation);
    return;
  }
  if (!result.authenticated) {
    options.setConversation(null);
    options.setError(
      result.messages?.map((message) => message.text).join(" ") || options.passwordError,
    );
    return;
  }
  await finishSignIn(options, options.destination);
};

/** Loads the issued session before navigating to authenticated content. */
const finishSignIn = async (options: LoginFormOptions, destination: string): Promise<void> => {
  const current = await loadMe();
  if (!current) throw new Error("Session unavailable.");
  options.onSuccess(current);
  options.navigate(destination, { replace: true });
};
