import { StandardDialog } from "@lite-nas/dialogs";
import { useLanguage } from "@lite-nas/language";
import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import type { AlertOccurrencesDialogLayoutProps } from "./types";

/**
 * Shared dialog shell used by alert occurrences chart and table variants.
 */
export const AlertOccurrencesDialogLayout = ({
  children,
  onClose,
  open,
  summary,
  title,
}: AlertOccurrencesDialogLayoutProps): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <StandardDialog
      data-testid="alert-occurrences-dialog"
      fullWidth
      maxWidth="lg"
      title={
        <Stack spacing={0.75}>
          <Typography variant="h3">{title}</Typography>
          <Typography color="text.secondary" variant="body2">
            {summary}
          </Typography>
        </Stack>
      }
      titleProps={{ "data-testid": "alert-occurrences-dialog-title" }}
      closeLabel={getLabel("dialogs.close")}
      onClose={onClose}
      open={open}
      contentProps={{ dividers: true }}
      actions={
        <Button data-testid="alert-occurrences-dialog-close-button" onClick={onClose}>
          {getLabel("admin.alerts.close")}
        </Button>
      }
    >
      {children}
    </StandardDialog>
  );
};
