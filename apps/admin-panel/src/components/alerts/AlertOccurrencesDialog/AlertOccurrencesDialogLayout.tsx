import Button from "@mui/material/Button";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogTitle from "@mui/material/DialogTitle";
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
  return (
    <Dialog
      data-testid="alert-occurrences-dialog"
      fullWidth
      maxWidth="lg"
      onClose={onClose}
      open={open}
    >
      <DialogTitle data-testid="alert-occurrences-dialog-title">
        <Stack spacing={0.75}>
          <Typography variant="h3">{title}</Typography>
          <Typography color="text.secondary" variant="body2">
            {summary}
          </Typography>
        </Stack>
      </DialogTitle>
      <DialogContent dividers>{children}</DialogContent>
      <DialogActions>
        <Button data-testid="alert-occurrences-dialog-close-button" onClick={onClose}>
          Close
        </Button>
      </DialogActions>
    </Dialog>
  );
};
