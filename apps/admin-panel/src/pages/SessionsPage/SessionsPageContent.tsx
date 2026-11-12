import type { SessionAction } from "@dto/management/sessions";
import { useSessions } from "@hooks/useSessions";
import { ConfirmActionDialog } from "@lite-nas/dialogs";
import { useLanguage } from "@lite-nas/language";
import ComputerIcon from "@mui/icons-material/Computer";
import DeleteIcon from "@mui/icons-material/Delete";
import KeyboardIcon from "@mui/icons-material/Keyboard";
import LockIcon from "@mui/icons-material/Lock";
import LockOpenIcon from "@mui/icons-material/LockOpen";
import TerminalIcon from "@mui/icons-material/Terminal";
import Button from "@mui/material/Button";
import IconButton from "@mui/material/IconButton";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Identifies the user-facing session category from loginctl properties. */
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The page keeps session metadata and its actions in one view.
const sessionKind = (session: ReturnType<typeof useSessions>["sessions"][number]): string => {
  if (session.class === "manager") return "manager";
  if (session.type === "wayland" || session.type === "x11") return "graphical";
  if (session.remote && session.service === "sshd") return "ssh";
  if (session.tty && session.tty !== "-") return "terminal";
  return "unknown";
};

/** Selects an accent color for the inferred session category. */
const sessionTone = (kind: string): string => {
  if (kind === "graphical") return "success.main";
  if (kind === "ssh") return "info.main";
  if (kind === "terminal") return "warning.main";
  return "text.disabled";
};

/** Selects a semantic icon color for the inferred session category. */
const sessionIconColor = (kind: string): "success" | "info" | "warning" | "disabled" => {
  if (kind === "graphical") return "success";
  if (kind === "ssh") return "info";
  if (kind === "terminal") return "warning";
  return "disabled";
};

/** Chooses an icon that matches the session category. */
const sessionIcon = (kind: string): typeof ComputerIcon => {
  if (kind === "ssh") return TerminalIcon;
  if (kind === "terminal") return KeyboardIcon;
  return ComputerIcon;
};

/** Chooses the action icon for a requested session operation. */
const actionIcon = (action: SessionAction): ReactElement => {
  switch (action) {
    case "lock":
      return <LockIcon />;
    case "unlock":
      return <LockOpenIcon />;
    case "terminate":
      return <DeleteIcon />;
  }
};

/** Renders the session page from its provider-owned state and actions. */
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The page keeps session metadata and its actions in one view.
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The page keeps session metadata and its actions in one view.
export const SessionsPageContent = (): ReactElement | null => {
  const { getLabel } = useLanguage();
  const {
    actionError,
    allowed,
    closeConfirmation,
    confirmation,
    isError,
    isLoading,
    page,
    pageCount,
    pending,
    requestAction,
    runAction,
    search,
    sessions,
    setPage,
    setSearch,
    totalSessions,
    visibleSessions,
  } = useSessions();
  const actions: SessionAction[] = ["lock", "unlock", "terminate"];

  if (!allowed) return null;

  return (
    <Stack data-testid="sessions-page" spacing={3}>
      <Typography variant="h1" data-testid="sessions-title">
        {getLabel("navigation.Loginctl")}
      </Typography>
      <Paper sx={{ p: 3 }}>
        <Stack spacing={2}>
          <Typography variant="h2">{getLabel("admin.sessions.title")}</Typography>
          <TextField
            data-testid="sessions-search"
            label={getLabel("admin.sessions.search")}
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
          <Typography data-testid="sessions-summary">
            {getLabel("admin.sessions.summary")
              .replace("{visible}", String(sessions.length))
              .replace("{total}", String(totalSessions))}
          </Typography>
          {isLoading && <Typography>{getLabel("admin.sessions.loading")}</Typography>}
          {isError && <Typography role="alert">{getLabel("admin.sessions.loadFailed")}</Typography>}
          {actionError && <Typography role="alert">{actionError}</Typography>}
          {!isLoading && !isError && sessions.length === 0 && (
            <Typography data-testid="sessions-empty">{getLabel("admin.sessions.empty")}</Typography>
          )}
          {visibleSessions.map((session) =>
            // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The page keeps session metadata and its actions in one view.
            (() => {
              const kind = sessionKind(session);
              const tone = sessionTone(kind);
              const SessionIcon = sessionIcon(kind);
              return (
                <Paper
                  key={session.id}
                  variant="outlined"
                  sx={{ p: 2, borderLeft: 4, borderLeftColor: tone }}
                  data-test-class="session-card"
                  data-test-name={session.id}
                >
                  <Stack
                    direction={{ xs: "column", md: "row" }}
                    justifyContent="space-between"
                    spacing={2}
                  >
                    <Stack spacing={0.5}>
                      <Stack alignItems="center" direction="row" spacing={1}>
                        <SessionIcon
                          aria-label={getLabel(`admin.sessions.kind.${kind}`)}
                          color={sessionIconColor(kind)}
                        />
                        <Typography color={tone} fontWeight={600}>
                          {getLabel(`admin.sessions.kind.${kind}`)}
                        </Typography>
                      </Stack>
                      <Typography variant="body1" fontWeight={600} color="primary.main">
                        {session.user}
                      </Typography>
                      <Stack direction="row" flexWrap="wrap" columnGap={2}>
                        <Typography variant="body2">
                          {getLabel("admin.sessions.id")}: {session.id} ·{" "}
                          {getLabel("admin.sessions.uid")}: {session.uid}
                        </Typography>
                        <Typography variant="body2">
                          {getLabel("admin.sessions.seat")}:{" "}
                          {session.seat === "-" || !session.seat ? "—" : session.seat} ·{" "}
                          {getLabel("admin.sessions.class")}: {session.class}
                          {" · "}
                          {getLabel("admin.sessions.type")}: {session.type || "—"}
                          {session.remote_host
                            ? ` · ${getLabel("admin.sessions.remoteHost")}: ${session.remote_host}`
                            : ""}
                        </Typography>
                        <Typography variant="body2">
                          {getLabel("admin.sessions.tty")}:{" "}
                          {session.tty === "-" || !session.tty ? "—" : session.tty} ·{" "}
                          {getLabel("admin.sessions.leader")}: {session.leader}
                        </Typography>
                      </Stack>
                    </Stack>
                    <Stack direction="row" spacing={1}>
                      {actions.map((action) => (
                        <IconButton
                          key={action}
                          data-testid={`session-${action}-${session.id}`}
                          aria-label={getLabel(`admin.sessions.${action}`)}
                          disabled={pending !== ""}
                          color={action === "terminate" ? "error" : "primary"}
                          onClick={() => requestAction(session, action)}
                        >
                          {actionIcon(action)}
                        </IconButton>
                      ))}
                    </Stack>
                  </Stack>
                </Paper>
              );
            })(),
          )}
          <ConfirmActionDialog
            open={confirmation !== null}
            title={getLabel("dialogs.areYouSure")}
            description={getLabel("admin.sessions.confirmAction")
              .replace(
                "{action}",
                confirmation ? getLabel(`admin.sessions.${confirmation.action}`) : "",
              )
              .replace("{session}", confirmation?.session.id ?? "")}
            confirmLabel={confirmation ? getLabel(`admin.sessions.${confirmation.action}`) : ""}
            requireYes={confirmation?.action === "terminate"}
            confirmColor={confirmation?.action === "terminate" ? "error" : "primary"}
            onClose={closeConfirmation}
            onConfirm={() =>
              confirmation
                ? runAction(confirmation.session, confirmation.action)
                : Promise.resolve()
            }
          />
          <Stack direction="row" alignItems="center" spacing={2}>
            <Button
              data-testid="sessions-previous"
              disabled={page <= 1}
              onClick={() => setPage(page - 1)}
            >
              {getLabel("admin.sessions.previous")}
            </Button>
            <Typography>
              {getLabel("admin.sessions.page")
                .replace("{page}", String(page))
                .replace("{pages}", String(pageCount))}
            </Typography>
            <Button
              data-testid="sessions-next"
              disabled={page >= pageCount}
              onClick={() => setPage(page + 1)}
            >
              {getLabel("admin.sessions.next")}
            </Button>
          </Stack>
        </Stack>
      </Paper>
    </Stack>
  );
};
