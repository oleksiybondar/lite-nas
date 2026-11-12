import CloseIcon from "@mui/icons-material/Close";
import Box from "@mui/material/Box";
import type { DialogProps } from "@mui/material/Dialog";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import type { DialogContentProps } from "@mui/material/DialogContent";
import DialogContent from "@mui/material/DialogContent";
import type { DialogTitleProps } from "@mui/material/DialogTitle";
import DialogTitle from "@mui/material/DialogTitle";
import IconButton from "@mui/material/IconButton";
import type { FormEventHandler, ReactElement, ReactNode } from "react";

/** Shared dialog shell that standardizes title, close control, content, and optional form actions. */
export function StandardDialog({
  title,
  closeLabel,
  closeDisabled = false,
  onClose,
  children,
  actions,
  onSubmit,
  contentProps,
  titleProps,
  showTitleBar = true,
  ...dialogProps
}: Omit<DialogProps, "title" | "onClose" | "children" | "onSubmit"> & {
  /** Shared dialog title. */
  title: ReactNode;
  /** Accessible label for the shared title-bar close control. */
  closeLabel: string;
  /** Disables the shared close control while a dialog action is in progress. */
  closeDisabled?: boolean;
  /** Handles close requests from Escape, backdrop, and the title-bar control. */
  onClose: () => void;
  /** Use-case-specific dialog content. */
  children: ReactNode;
  /** Optional use-case-specific footer controls. */
  actions?: ReactNode;
  /** Optional native form submit handler for dialog actions and Enter submission. */
  onSubmit?: FormEventHandler<HTMLFormElement>;
  /** Props for the shared content region. */
  contentProps?: DialogContentProps;
  /** MUI title props for selectors and use-case-specific title styling. */
  titleProps?: DialogTitleProps & { "data-testid"?: string };
  /** Hides the shared title bar for immersive dialogs with an in-content toolbar. */
  showTitleBar?: boolean;
}): ReactElement {
  const header = (
    <DialogTitle
      {...titleProps}
      sx={{ display: "flex", alignItems: "center", justifyContent: "space-between", gap: 1 }}
    >
      {title}
      <IconButton
        aria-label={closeLabel}
        onClick={onClose}
        disabled={closeDisabled}
        edge="end"
        size="small"
      >
        <CloseIcon />
      </IconButton>
    </DialogTitle>
  );
  const titleBar = showTitleBar ? header : null;
  const content = <DialogContent {...contentProps}>{children}</DialogContent>;
  const footer = actions ? <DialogActions>{actions}</DialogActions> : null;

  return (
    <Dialog {...dialogProps} onClose={() => onClose()}>
      {onSubmit ? (
        <Box
          component="form"
          onSubmit={onSubmit}
          sx={{ display: "flex", flexDirection: "column", minHeight: 0 }}
        >
          {titleBar}
          {content}
          {footer}
        </Box>
      ) : (
        <>
          {titleBar}
          {content}
          {footer}
        </>
      )}
    </Dialog>
  );
}
