import { useLanguage } from "@lite-nas/language";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Reports sensor loading and profile save failures above the sensor sections. */
export const SensorPageStatus = ({
  isLoading,
  isError,
  saveError,
}: {
  isLoading: boolean;
  isError: boolean;
  saveError: boolean;
}): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <>
      {isLoading && <Typography>{getLabel("common.loading")}</Typography>}
      {isError && <Typography color="error">{getLabel("admin.sensors.loadFailed")}</Typography>}
      {saveError && (
        <Typography color="error" role="alert">
          {getLabel("admin.sensors.saveFailed")}
        </Typography>
      )}
    </>
  );
};
