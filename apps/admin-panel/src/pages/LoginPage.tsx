import { PublicAppLayout } from "@components/layout/PublicAppLayout";
import { useAuth } from "@hooks/useAuth";
import { authDestination, loginDestination } from "@lite-nas/profile";
import Button from "@mui/material/Button";
import Container from "@mui/material/Container";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import { authSessionResponseSchema } from "@schemas/auth/session";
import type { FormEvent, ReactElement } from "react";
import { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";

/**
 * Browser login page rendered when the auth guard resolves an anonymous state.
 */
export const LoginPage = (): ReactElement => {
  const { login } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const from = (location.state as { from?: unknown } | null)?.from;
  const destination = loginDestination(location.search, from);
  const { error, isSubmitting, onSubmit, password, setPassword, setUsername, username } =
    useLoginForm(login, destination, (returned) => {
      void navigate(authDestination(returned ?? destination), { replace: true });
    });

  return (
    <PublicAppLayout>
      <Container data-testid="login-page" maxWidth="xs" sx={{ py: 8 }}>
        <Paper data-testid="login-card" sx={{ p: 4 }}>
          <Stack component="form" data-testid="login-form" onSubmit={onSubmit} spacing={3}>
            <Stack spacing={1}>
              <Typography data-testid="login-title" variant="h1">
                Sign in
              </Typography>
            </Stack>
            <TextField
              autoComplete="username"
              data-testid="login-username-field"
              label="Login"
              name="login"
              onChange={(event) => {
                setUsername(event.target.value);
              }}
              required
              value={username}
            />
            <TextField
              autoComplete="current-password"
              data-testid="login-password-field"
              label="Password"
              name="password"
              onChange={(event) => {
                setPassword(event.target.value);
              }}
              required
              type="password"
              value={password}
            />
            {error !== null ? (
              <Typography color="error" data-testid="login-error-message">
                {error}
              </Typography>
            ) : null}
            <Button
              data-testid="login-submit-button"
              disabled={isSubmitting}
              type="submit"
              variant="contained"
            >
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
const useLoginForm = (
  login: LoginAction,
  destination: string,
  onAuthenticated: (redirectTo?: string) => void,
) => {
  const [error, setError] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [password, setPassword] = useState("");
  const [username, setUsername] = useState("");

  const onSubmit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault();
    setError(null);
    setIsSubmitting(true);

    try {
      const response = await login(username, password, destination);
      if (!response.ok) {
        setError("Login failed.");
      } else {
        onAuthenticated(await readLoginRedirect(response));
      }
    } catch {
      setError("Login failed.");
    } finally {
      setIsSubmitting(false);
    }
  };

  return { error, isSubmitting, onSubmit, password, setPassword, setUsername, username };
};

/** Reads the echoed path while preserving login links across older gateway deployments. */
const readLoginRedirect = async (response: Response): Promise<string | undefined> => {
  try {
    return authSessionResponseSchema.parse(await response.clone().json()).data.redirect_to;
  } catch {
    return undefined;
  }
};
