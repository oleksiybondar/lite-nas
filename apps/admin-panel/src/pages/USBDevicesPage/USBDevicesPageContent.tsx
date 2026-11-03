import { useUSBDevices } from "@hooks/useUSBDevices";
import { ConfirmActionDialog } from "@lite-nas/dialogs";
import { useLanguage } from "@lite-nas/language";
import BlockIcon from "@mui/icons-material/Block";
import CheckCircleOutlineIcon from "@mui/icons-material/CheckCircleOutline";
import Button from "@mui/material/Button";
import Checkbox from "@mui/material/Checkbox";
import FormControlLabel from "@mui/material/FormControlLabel";
import IconButton from "@mui/material/IconButton";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

/** Maps the USBGuard authorization state to its accessible card accent. */
const deviceTone = (state: string): string => {
  if (state === "allow") return "success.main";
  if (state === "block") return "error.main";
  return "warning.main";
};

/** Renders the USBGuard page from its provider-owned state and actions. */
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The page keeps device state and authorization actions in one view.
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The page keeps device state and authorization actions in one view.
export const USBDevicesPageContent = (): ReactElement | null => {
  const { getLabel } = useLanguage();
  const {
    actionError,
    allowed,
    closeConfirmation,
    confirmation,
    devices,
    isError,
    isLoading,
    page,
    pageCount,
    pending,
    permanent,
    requestAction,
    runAction,
    search,
    setPage,
    setPermanent,
    setSearch,
    totalDevices,
    visibleDevices,
  } = useUSBDevices();
  if (!allowed) return null;

  return (
    <Stack data-testid="usb-devices-page" spacing={3}>
      <Typography variant="h1" data-testid="usb-devices-title">
        {getLabel("navigation.USBDevices")}
      </Typography>
      <Paper sx={{ p: 3 }}>
        <Stack spacing={2}>
          <Typography variant="h2">{getLabel("admin.usbDevices.title")}</Typography>
          <TextField
            data-testid="usb-devices-search"
            label={getLabel("admin.usbDevices.search")}
            value={search}
            onChange={(event) => setSearch(event.target.value)}
          />
          <FormControlLabel
            control={
              <Checkbox
                data-testid="usb-devices-permanent"
                checked={permanent}
                onChange={(event) => setPermanent(event.target.checked)}
              />
            }
            label={getLabel("admin.usbDevices.permanent")}
          />
          <Typography data-testid="usb-devices-summary">
            {getLabel("admin.usbDevices.summary")
              .replace("{visible}", String(devices.length))
              .replace("{total}", String(totalDevices))}
          </Typography>
          {isLoading && <Typography>{getLabel("admin.usbDevices.loading")}</Typography>}
          {isError && (
            <Typography role="alert">{getLabel("admin.usbDevices.loadFailed")}</Typography>
          )}
          {actionError && <Typography role="alert">{actionError}</Typography>}
          {!isLoading && !isError && devices.length === 0 && (
            <Typography data-testid="usb-devices-empty">
              {getLabel("admin.usbDevices.empty")}
            </Typography>
          )}
          {visibleDevices.map((device) => {
            const tone = deviceTone(device.state);
            return (
              <Paper
                key={device.id}
                variant="outlined"
                sx={{ p: 2, borderLeft: 4, borderLeftColor: tone }}
                data-test-class="usb-device-card"
                data-test-name={device.id}
              >
                <Stack
                  direction={{ xs: "column", md: "row" }}
                  justifyContent="space-between"
                  spacing={2}
                >
                  <Stack spacing={0.5}>
                    <Typography fontWeight={600} color="primary.main">
                      {device.name || getLabel("admin.usbDevices.unnamed")}
                    </Typography>
                    <Typography variant="body2">
                      {getLabel("admin.usbDevices.id")}: {device.id} ·{" "}
                      {getLabel("admin.usbDevices.state")}:{" "}
                      {getLabel(`admin.usbDevices.states.${device.state}`)}
                    </Typography>
                    {device.product && (
                      <Typography variant="body2">
                        {getLabel("admin.usbDevices.product")}: {device.product}
                      </Typography>
                    )}
                    <Typography variant="body2">
                      {getLabel("admin.usbDevices.hash")}: {device.hash}
                    </Typography>
                    {device.attributes && (
                      <Typography variant="body2" sx={{ overflowWrap: "anywhere" }}>
                        {getLabel("admin.usbDevices.attributes")}: {device.attributes}
                      </Typography>
                    )}
                  </Stack>
                  <Stack direction="row" spacing={1}>
                    <IconButton
                      data-testid={`usb-device-allow-${device.id}`}
                      aria-label={getLabel("admin.usbDevices.allow")}
                      disabled={pending !== "" || device.state === "allow"}
                      color="success"
                      onClick={() => requestAction(device, "allow")}
                    >
                      <CheckCircleOutlineIcon />
                    </IconButton>
                    <IconButton
                      data-testid={`usb-device-block-${device.id}`}
                      aria-label={getLabel("admin.usbDevices.block")}
                      disabled={pending !== "" || device.state === "block"}
                      color="error"
                      onClick={() => requestAction(device, "block")}
                    >
                      <BlockIcon />
                    </IconButton>
                  </Stack>
                </Stack>
              </Paper>
            );
          })}
          <ConfirmActionDialog
            open={confirmation !== null}
            title={getLabel("dialogs.areYouSure")}
            description={getLabel("admin.usbDevices.confirmAction")
              .replace(
                "{action}",
                confirmation ? getLabel(`admin.usbDevices.${confirmation.action}`) : "",
              )
              .replace("{device}", confirmation?.device.name || confirmation?.device.id || "")}
            confirmLabel={confirmation ? getLabel(`admin.usbDevices.${confirmation.action}`) : ""}
            requireYes={permanent}
            confirmColor={confirmation?.action === "block" ? "error" : "primary"}
            onClose={closeConfirmation}
            onConfirm={() =>
              confirmation ? runAction(confirmation.device, confirmation.action) : Promise.resolve()
            }
          />
          <Stack direction="row" alignItems="center" spacing={2}>
            <Button
              data-testid="usb-devices-previous"
              disabled={page <= 1}
              onClick={() => setPage(page - 1)}
            >
              {getLabel("admin.usbDevices.previous")}
            </Button>
            <Typography>
              {getLabel("admin.usbDevices.page")
                .replace("{page}", String(page))
                .replace("{pages}", String(pageCount))}
            </Typography>
            <Button
              data-testid="usb-devices-next"
              disabled={page >= pageCount}
              onClick={() => setPage(page + 1)}
            >
              {getLabel("admin.usbDevices.next")}
            </Button>
          </Stack>
        </Stack>
      </Paper>
    </Stack>
  );
};
