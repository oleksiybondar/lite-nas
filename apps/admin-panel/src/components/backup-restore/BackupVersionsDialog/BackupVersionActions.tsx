import { useLanguage } from "@lite-nas/language";
import CallMergeIcon from "@mui/icons-material/CallMerge";
import CompareArrowsIcon from "@mui/icons-material/CompareArrows";
import DeleteIcon from "@mui/icons-material/Delete";
import RestoreIcon from "@mui/icons-material/Restore";
import IconButton from "@mui/material/IconButton";
import Stack from "@mui/material/Stack";
import Tooltip from "@mui/material/Tooltip";
import type { ReactElement } from "react";

type VersionAction = {
  name: "compare" | "restore" | "merge" | "delete";
  labelKey: string;
  icon: typeof CompareArrowsIcon;
  disabled: boolean;
  run: () => void;
};

/** Renders the compare, restore, merge, and delete actions for one version. */
export const BackupVersionActions = ({
  version,
  busy,
  sessionActive,
  onCompare,
  onRestore,
  onMerge,
  onDelete,
}: {
  version: string;
  busy: boolean;
  sessionActive: boolean;
  onCompare: () => void;
  onRestore: () => void;
  onMerge: () => void;
  onDelete: () => void;
}): ReactElement => {
  const { getLabel } = useLanguage();
  const actions: VersionAction[] = [
    {
      name: "compare",
      labelKey: "admin.backup.compare",
      icon: CompareArrowsIcon,
      disabled: busy,
      run: onCompare,
    },
    {
      name: "restore",
      labelKey: "admin.backup.restore",
      icon: RestoreIcon,
      disabled: busy || sessionActive,
      run: onRestore,
    },
    {
      name: "merge",
      labelKey: "admin.backup.startMerge",
      icon: CallMergeIcon,
      disabled: busy || sessionActive,
      run: onMerge,
    },
    {
      name: "delete",
      labelKey: "admin.backup.delete",
      icon: DeleteIcon,
      disabled: busy || sessionActive,
      run: onDelete,
    },
  ];
  return (
    <Stack direction="row" spacing={0.5}>
      {actions.map(({ name, labelKey, icon: Icon, disabled, run }) => (
        <Tooltip key={name} title={getLabel(labelKey)}>
          <span>
            <IconButton
              data-testid={`backup-${name}-${version}`}
              aria-label={getLabel(labelKey)}
              color={name === "delete" ? "error" : "default"}
              disabled={disabled}
              onClick={run}
            >
              <Icon />
            </IconButton>
          </span>
        </Tooltip>
      ))}
    </Stack>
  );
};
