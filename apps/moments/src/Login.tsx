import Button from "@mui/material/Button";
import Container from "@mui/material/Container";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { FormEvent, ReactElement } from "react";
import { useState } from "react";
import { loadMe, login } from "./api";
import type { Me } from "./schemas";

/** Collects credentials for the existing LiteNAS browser login flow. */
export const Login = ({ onSuccess }: { onSuccess: (me: Me) => void }): ReactElement => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const onSubmit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault();
    try {
      await login({ login: username, password });
      const current = await loadMe();
      if (!current) throw new Error("Session unavailable.");
      onSuccess(current);
    } catch {
      setError("Sign in failed.");
    }
  };
  return (
    <Container data-testid="moments-login" maxWidth="sm" sx={{ py: 8 }}>
      <Paper sx={{ p: 4 }}>
        <Stack component="form" onSubmit={(event) => void onSubmit(event)} spacing={2}>
          <Typography variant="h1">Sign in to Moments</Typography>
          <input
            aria-label="Login"
            autoComplete="username"
            data-testid="moments-login-name"
            onChange={(event) => setUsername(event.target.value)}
            value={username}
          />
          <input
            aria-label="Password"
            autoComplete="current-password"
            data-testid="moments-login-password"
            onChange={(event) => setPassword(event.target.value)}
            type="password"
            value={password}
          />
          {error && <Typography color="error">{error}</Typography>}
          <Button data-testid="moments-login-submit" type="submit" variant="contained">
            Sign in
          </Button>
        </Stack>
      </Paper>
    </Container>
  );
};
