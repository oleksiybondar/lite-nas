import { PublicAppLayout } from "@components/layout/PublicAppLayout";
import { useAuth } from "@hooks/useAuth";
import Button from "@mui/material/Button";
import Container from "@mui/material/Container";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { FormEvent, ReactElement } from "react";
import { useState } from "react";

/**
 * Browser login page rendered when the auth guard resolves an anonymous state.
 */
export const LoginPage = (): ReactElement => {
  const { login } = useAuth();
  const { error, isSubmitting, onSubmit, password, setPassword, setUsername, username } =
    useLoginForm(login);

  return (
    <PublicAppLayout>
      <Container maxWidth="xs" sx={{ py: 8 }}>
        <Paper sx={{ p: 4 }}>
          <Stack component="form" onSubmit={onSubmit} spacing={3}>
            <Stack spacing={1}>
              <Typography variant="h1">Sign in</Typography>
            </Stack>
            <TextField
              autoComplete="username"
              label="Login"
              onChange={(event) => {
                setUsername(event.target.value);
              }}
              required
              value={username}
            />
            <TextField
              autoComplete="current-password"
              label="Password"
              onChange={(event) => {
                setPassword(event.target.value);
              }}
              required
              type="password"
              value={password}
            />
            {error !== null ? <Typography color="error">{error}</Typography> : null}
            <Button disabled={isSubmitting} type="submit" variant="contained">
              Sign in
            </Button>
          </Stack>
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
const useLoginForm = (login: LoginAction) => {
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [password, setPassword] = useState("");
  const [username, setUsername] = useState("");

  const onSubmit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      const response = await login(username, password);
      if (!response.ok) {
        setError("Login failed.");
      }
    } catch {
      setError("Login failed.");
    } finally {
      setIsSubmitting(false);
    }
  };

  return { error, isSubmitting, onSubmit, password, setPassword, setUsername, username };
};
