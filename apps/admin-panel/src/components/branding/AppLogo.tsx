import StorageRoundedIcon from "@mui/icons-material/StorageRounded";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

export const AppLogo = (): ReactElement => {
  return (
    <Stack alignItems="center" direction="row" spacing={1.25}>
      <StorageRoundedIcon color="primary" fontSize="small" />
      <Typography component="span" fontWeight={700} variant="subtitle1">
        LiteNAS
      </Typography>
    </Stack>
  );
};
