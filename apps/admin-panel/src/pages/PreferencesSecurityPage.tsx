import { useAuth } from "@hooks/useAuth";
import { authFlowSchema, PasswordChangePanel } from "@lite-nas/authentication";
import { useLanguage } from "@lite-nas/language";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Shows account security controls for the signed-in admin-panel user. */
export const PreferencesSecurityPage = (): ReactElement => {
  const { getLabel } = useLanguage();
  const { startPasswordChange, replyPasswordChange } = useAuth();

  return (
    <Stack data-testid="preferences-security-page" spacing={3}>
      <Typography variant="h1">{getLabel("navigation.Security")}</Typography>
      <Typography color="text.secondary">{getLabel("admin.security.summary")}</Typography>
      <PasswordChangePanel
        onStart={async () => {
          if (!startPasswordChange) throw new Error("Password change unavailable.");
          const response = await startPasswordChange();
          if (!response.ok) throw new Error("Password change failed.");
          return authFlowSchema.parse(await response.json()).data;
        }}
        onReply={async (conversation, answer) => {
          if (!replyPasswordChange) throw new Error("Password change unavailable.");
          const response = await replyPasswordChange(conversation, answer);
          if (!response.ok) throw new Error("Password change failed.");
          return authFlowSchema.parse(await response.json()).data;
        }}
      />
    </Stack>
  );
};
