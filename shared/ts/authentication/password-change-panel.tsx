import { useLanguage } from "@lite-nas/language";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import { type FormEvent, type ReactElement, useState } from "react";
import { PasswordPromptForm } from "./password-prompt-form";
import type { AuthFlow, PasswordConversation } from "./schemas";

/** Runs an on-demand PAM password conversation while the session stays active. */
export const PasswordChangePanel = ({
  onStart,
  onReply,
}: {
  onStart: () => Promise<AuthFlow>;
  onReply: (conversation: PasswordConversation, answer: string) => Promise<AuthFlow>;
}): ReactElement => {
  const { getLabel } = useLanguage();
  const [conversation, setConversation] = useState<PasswordConversation | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const apply = (flow: AuthFlow): void => {
    setConversation(flow.conversation ?? null);
    if (flow.status === "denied") {
      setError(passwordChangeError(flow, getLabel("common.passwordChangeFailed")));
      return;
    }
    if (flow.status === "authenticated") setSuccess(!flow.conversation);
  };
  const start = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault();
    setBusy(true);
    setError(null);
    setSuccess(false);
    try {
      apply(await onStart());
    } catch {
      setError(getLabel("common.passwordChangeFailed"));
    } finally {
      setBusy(false);
    }
  };
  const reply = async (answer: string): Promise<void> => {
    if (!conversation) return;
    setBusy(true);
    setError(null);
    try {
      apply(await onReply(conversation, answer));
    } catch {
      setError(getLabel("common.passwordChangeFailed"));
    } finally {
      setBusy(false);
    }
  };

  if (conversation)
    return (
      <PasswordPromptForm conversation={conversation} busy={busy} error={error} onAnswer={reply} />
    );
  return (
    <Stack component="form" onSubmit={(event) => void start(event)} spacing={2} sx={{ mt: 3 }}>
      <Typography variant="h6">{getLabel("common.passwordChange")}</Typography>
      {error ? <Typography color="error">{error}</Typography> : null}
      {success ? (
        <Typography color="success.main">{getLabel("common.passwordChanged")}</Typography>
      ) : null}
      <Button data-testid="password-change-start" disabled={busy} type="submit" variant="contained">
        {getLabel("common.passwordChange")}
      </Button>
    </Stack>
  );
};

/** Returns PAM's explanation when a password change is denied. */
const passwordChangeError = (flow: AuthFlow, fallback: string): string =>
  flow.messages?.map((message) => message.text).join(" ") || fallback;
