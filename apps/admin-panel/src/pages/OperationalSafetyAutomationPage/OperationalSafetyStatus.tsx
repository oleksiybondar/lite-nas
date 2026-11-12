import { useLanguage } from "@lite-nas/language";
import { Alert, Typography } from "@mui/material";
import type { ReactElement } from "react";

/** Reports loading, request failures, and the latest operator outcome. */
export const OperationalSafetyStatus = ({
  loading,
  error,
  actionError,
  outcome,
}: {
  loading: boolean;
  error: boolean;
  actionError: boolean;
  outcome: string;
}): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <>
      {error && <Alert severity="error">{getLabel("admin.operationalSafety.loadFailed")}</Alert>}
      {actionError && (
        <Alert severity="error">{getLabel("admin.operationalSafety.actionFailed")}</Alert>
      )}
      {outcome && <Alert severity="info">{outcome}</Alert>}
      {loading && <Typography>{getLabel("admin.operationalSafety.loading")}</Typography>}
    </>
  );
};
