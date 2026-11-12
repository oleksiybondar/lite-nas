import { useLanguage } from "@lite-nas/language";
import RefreshIcon from "@mui/icons-material/Refresh";
import { Button, Stack, TextField } from "@mui/material";
import type { ReactElement } from "react";

/** Searches the automation inventory and refreshes it on request. */
export const OperationalSafetyToolbar = ({
  search,
  onSearch,
  onRefresh,
}: {
  search: string;
  onSearch: (value: string) => void;
  onRefresh: () => void;
}): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Stack direction={{ xs: "column", sm: "row" }} spacing={1}>
      <TextField
        fullWidth
        label={getLabel("admin.operationalSafety.search")}
        value={search}
        onChange={(event) => onSearch(event.target.value)}
      />
      <Button startIcon={<RefreshIcon />} onClick={onRefresh}>
        {getLabel("admin.operationalSafety.refresh")}
      </Button>
    </Stack>
  );
};
