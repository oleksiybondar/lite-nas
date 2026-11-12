import { useLanguage } from "@lite-nas/language";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import { type FormEvent, type ReactElement, useState } from "react";
import type { PasswordConversation } from "./schemas";

/** Displays the current PAM prompt and submits exactly one answer. */
export const PasswordPromptForm = ({
  conversation,
  busy,
  onAnswer,
  error,
}: {
  conversation: PasswordConversation;
  busy: boolean;
  onAnswer: (answer: string) => Promise<void>;
  error?: string | null;
}): ReactElement => {
  const { getLabel } = useLanguage();
  const [answer, setAnswer] = useState("");
  const submit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault();
    await onAnswer(answer);
    setAnswer("");
  };
  return (
    <Stack component="form" onSubmit={(event) => void submit(event)} spacing={2}>
      <Typography variant="h1">{getLabel("common.passwordChange")}</Typography>
      {keyMessages(conversation.messages)?.map(({ key, message }) => (
        <Typography key={key} color={message.level === "error" ? "error" : "text.primary"}>
          {message.text}
        </Typography>
      ))}
      <TextField
        autoFocus
        data-testid="pam-prompt-answer"
        label={conversation.prompt || getLabel("common.passwordPrompt")}
        onChange={(event) => setAnswer(event.target.value)}
        required
        type={conversation.secret ? "password" : "text"}
        value={answer}
      />
      {error ? <Typography color="error">{error}</Typography> : null}
      <Button data-testid="pam-prompt-submit" disabled={busy} type="submit" variant="contained">
        {getLabel("common.continue")}
      </Button>
    </Stack>
  );
};

/** Gives repeated PAM messages stable keys without relying on array positions. */
const keyMessages = (messages: PasswordConversation["messages"]) => {
  const counts = new Map<string, number>();
  return messages?.map((message) => {
    const content = `${message.level}:${message.text}`;
    const occurrence = counts.get(content) ?? 0;
    counts.set(content, occurrence + 1);
    return { key: `${content}:${occurrence}`, message };
  });
};
