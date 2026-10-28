import { useLanguage } from "@lite-nas/language";
import { loginDestination } from "@lite-nas/profile";
import Button from "@mui/material/Button";
import Container from "@mui/material/Container";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { FormEvent, ReactElement } from "react";
import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { loadMe, login } from "./api";
import { MomentsLayout } from "./MomentsLayout";
import { MomentsPublicHeader } from "./MomentsPublicHeader";
import type { Me } from "./schemas";

/** Collects credentials for the existing LiteNAS browser login flow. */
export const Login = ({ onSuccess }: { onSuccess: (me: Me) => void }): ReactElement => {
  const { getLabel } = useLanguage();
  const location = useLocation();
  const navigate = useNavigate();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const onSubmit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);
    try {
      const requested = loginDestination(location.search);
      const destination = await login({ login: username, password }, requested);
      const current = await loadMe();
      if (!current) throw new Error("Session unavailable.");
      onSuccess(current);
      navigate(destination, { replace: true });
    } catch {
      setError(getLabel("common.signInFailed"));
    } finally {
      setIsSubmitting(false);
    }
  };
  return (
    <MomentsLayout header={<MomentsPublicHeader />}>
      <Container data-testid="moments-login" maxWidth="xs" sx={{ py: 8, my: "auto" }}>
        <Paper sx={{ p: { xs: 3, sm: 4 } }}>
          <LoginForm
            error={error}
            isSubmitting={isSubmitting}
            onSubmit={onSubmit}
            setPassword={setPassword}
            setUsername={setUsername}
            username={username}
            password={password}
          />
        </Paper>
      </Container>
    </MomentsLayout>
  );
};

type LoginFormProps = {
  error: string | null;
  isSubmitting: boolean;
  onSubmit: (event: FormEvent<HTMLFormElement>) => Promise<void>;
  setPassword: (password: string) => void;
  setUsername: (username: string) => void;
  username: string;
  password: string;
};

/** Renders the credential fields and submits them through the parent login flow. */
const LoginForm = ({
  error,
  isSubmitting,
  onSubmit,
  setPassword,
  setUsername,
  username,
  password,
}: LoginFormProps): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Stack component="form" onSubmit={(event) => void onSubmit(event)} spacing={3}>
      <Typography variant="h1">{getLabel("moments.signIn")}</Typography>
      <TextField
        autoComplete="username"
        data-testid="moments-login-name"
        fullWidth
        label={getLabel("common.login")}
        name="login"
        onChange={(event) => setUsername(event.target.value)}
        required
        value={username}
      />
      <TextField
        autoComplete="current-password"
        data-testid="moments-login-password"
        fullWidth
        label={getLabel("common.password")}
        name="password"
        onChange={(event) => setPassword(event.target.value)}
        required
        type="password"
        value={password}
      />
      {error && <Typography color="error">{error}</Typography>}
      <Button
        data-testid="moments-login-submit"
        disabled={isSubmitting}
        type="submit"
        variant="contained"
      >
        {getLabel("common.signIn")}
      </Button>
    </Stack>
  );
};
