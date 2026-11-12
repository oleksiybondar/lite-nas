import { usePollingQuery } from "@domain/monitoring/hooks/usePollingQuery";
import { useApi } from "@hooks/useApi";
import { useRbac } from "@hooks/useRbac";
import { ConfirmActionDialog } from "@lite-nas/dialogs";
import { useLanguage } from "@lite-nas/language";
import AppsRoundedIcon from "@mui/icons-material/AppsRounded";
import DeleteIcon from "@mui/icons-material/Delete";
import HelpOutlineIcon from "@mui/icons-material/HelpOutline";
import PhotoLibraryIcon from "@mui/icons-material/PhotoLibrary";
import IconButton from "@mui/material/IconButton";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import {
  type AuthSession,
  authSessionRevokeResponseSchema,
  authSessionsResponseSchema,
} from "@schemas/management/auth-sessions";
import { useQueryClient } from "@tanstack/react-query";
import type { ReactElement } from "react";
import { useState } from "react";

const authSessionsQueryKey = ["auth-sessions"];

/** Lists active LiteNAS Auth sessions and lets administrators revoke them. */
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Session state, authorization, and visible status are kept in this single page.
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The inventory, revoke action, and confirmation are one reviewable flow.
export const AuthSessionsPage = (): ReactElement | null => {
  const { getLabel } = useLanguage();
  const { requireAdmin } = useRbac();
  const { post } = useApi();
  const queryClient = useQueryClient();
  const allowed = requireAdmin();
  const [selected, setSelected] = useState<AuthSession | null>(null);
  const [pending, setPending] = useState(false);
  const [actionError, setActionError] = useState("");
  const { data, isLoading, isError } = usePollingQuery<AuthSession[]>({
    enabled: allowed,
    errorMessage: getLabel("admin.authSessions.loadFailed"),
    parser: (response) => authSessionsResponseSchema.parse(response).data,
    path: "/api/auth/sessions",
    queryKey: authSessionsQueryKey,
  });

  if (!allowed) return null;

  // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The revoke flow handles transport, confirmation, and query refresh together.
  const revoke = async (): Promise<void> => {
    if (!selected || pending) return;
    setPending(true);
    setActionError("");
    try {
      const response = await post(
        `/api/auth/sessions/${encodeURIComponent(selected.session_id)}/revoke`,
      ).execute();
      if (!response.ok) throw new Error("revoke failed");
      const result = authSessionRevokeResponseSchema.parse(await response.json());
      if (!result.revoked) throw new Error("session no longer active");
      setSelected(null);
      await queryClient.invalidateQueries({ queryKey: authSessionsQueryKey });
    } catch {
      setActionError(getLabel("admin.authSessions.revokeFailed"));
    } finally {
      setPending(false);
    }
  };

  return (
    <Stack data-testid="auth-sessions-page" spacing={3}>
      <Typography variant="h1" data-testid="auth-sessions-title">
        {getLabel("navigation.LiteNAS Auth")}
      </Typography>
      <Paper sx={{ p: 3 }}>
        <Stack spacing={2}>
          <Typography>{getLabel("admin.authSessions.description")}</Typography>
          {isLoading && <Typography>{getLabel("admin.authSessions.loading")}</Typography>}
          {isError && (
            <Typography role="alert">{getLabel("admin.authSessions.loadFailed")}</Typography>
          )}
          {actionError && <Typography role="alert">{actionError}</Typography>}
          {!isLoading && !isError && data?.length === 0 && (
            <Typography data-testid="auth-sessions-empty">
              {getLabel("admin.authSessions.empty")}
            </Typography>
          )}
          {data?.map((session) => (
            <Paper
              key={session.session_id}
              variant="outlined"
              sx={{ p: 2, borderLeft: 4, borderLeftColor: "primary.main" }}
              data-test-class="auth-session-card"
              data-test-name={session.session_id}
            >
              <Stack
                direction={{ xs: "column", sm: "row" }}
                justifyContent="space-between"
                spacing={2}
              >
                <Stack spacing={0.5}>
                  <Stack direction="row" alignItems="center" spacing={0.75}>
                    {session.source === "moments" && (
                      <PhotoLibraryIcon color="secondary" fontSize="small" />
                    )}
                    {session.source === "admin-panel" && (
                      <AppsRoundedIcon color="primary" fontSize="small" />
                    )}
                    {!session.source && <HelpOutlineIcon color="disabled" fontSize="small" />}
                    <Typography color="primary.main" fontWeight={600}>
                      {session.login}
                    </Typography>
                  </Stack>
                  <Typography
                    variant="body2"
                    sx={{
                      alignItems: "center",
                      display: "flex",
                      gap: 0.5,
                      overflowX: "auto",
                      whiteSpace: "nowrap",
                    }}
                  >
                    {getLabel("admin.authSessions.subject")}: {session.subject} ·{" "}
                    {getLabel("admin.authSessions.id")}: {session.session_id} ·{" "}
                    {getLabel("admin.authSessions.expiresAt")}:{" "}
                    {new Date(session.expires_at).toLocaleString()} ·
                    {getLabel(`admin.authSessions.source.${session.source ?? "unknown"}`)}
                  </Typography>
                </Stack>
                <IconButton
                  color="error"
                  disabled={pending}
                  aria-label={getLabel("admin.authSessions.revoke")}
                  onClick={() => setSelected(session)}
                  data-test-class="auth-session-revoke"
                  data-test-name={session.session_id}
                >
                  <DeleteIcon />
                </IconButton>
              </Stack>
            </Paper>
          ))}
        </Stack>
      </Paper>
      <ConfirmActionDialog
        open={selected !== null}
        title={getLabel("dialogs.areYouSure")}
        description={getLabel("admin.authSessions.confirmRevoke").replace(
          "{login}",
          selected?.login ?? "",
        )}
        confirmLabel={getLabel("admin.authSessions.revoke")}
        confirmColor="error"
        requireYes
        onClose={() => setSelected(null)}
        onConfirm={revoke}
      />
    </Stack>
  );
};
