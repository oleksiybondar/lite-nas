import { useLanguage } from "@lite-nas/language";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { FormEvent, ReactElement } from "react";
import { PasswordPromptForm } from "./password-prompt-form";
import type { PasswordConversation } from "./schemas";

/** Renders login or the current PAM prompt from the browser auth state. */
export const AuthenticationForm = ({
  title,
  username,
  password,
  setUsername,
  setPassword,
  onSubmit,
  onAnswer,
  conversation,
  busy,
  error,
  testIds,
}: {
  title: string;
  username: string;
  password: string;
  setUsername: (value: string) => void;
  setPassword: (value: string) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => Promise<void>;
  onAnswer: (answer: string) => Promise<void>;
  conversation: PasswordConversation | null;
  busy: boolean;
  error: string | null;
  testIds: {
    form?: string;
    title?: string;
    username: string;
    password: string;
    submit: string;
    error?: string;
  };
}): ReactElement => {
  const { getLabel } = useLanguage();
  if (conversation)
    return (
      <PasswordPromptForm
        conversation={conversation}
        busy={busy}
        error={error}
        onAnswer={onAnswer}
      />
    );
  return (
    <Stack
      component="form"
      data-testid={testIds.form}
      onSubmit={(event) => void onSubmit(event)}
      spacing={3}
    >
      <Typography data-testid={testIds.title} variant="h1">
        {title}
      </Typography>
      <TextField
        autoComplete="username"
        data-testid={testIds.username}
        label={getLabel("common.login")}
        name="login"
        onChange={(event) => setUsername(event.target.value)}
        required
        value={username}
      />
      <TextField
        autoComplete="current-password"
        data-testid={testIds.password}
        label={getLabel("common.password")}
        name="password"
        onChange={(event) => setPassword(event.target.value)}
        required
        type="password"
        value={password}
      />
      {error ? (
        <Typography color="error" data-testid={testIds.error}>
          {error}
        </Typography>
      ) : null}
      <Button data-testid={testIds.submit} disabled={busy} type="submit" variant="contained">
        {getLabel("common.signIn")}
      </Button>
    </Stack>
  );
};
