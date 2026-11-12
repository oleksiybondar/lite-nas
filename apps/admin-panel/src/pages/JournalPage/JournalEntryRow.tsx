import { useLanguage } from "@lite-nas/language";
import { Paper, Stack, Typography } from "@mui/material";
import type { JournalEntry } from "@schemas/management/journal";
import type { ReactElement, ReactNode } from "react";

/** Marks literal matches in a journal message without injecting HTML. */
const highlight = (message: string, search: string): ReactNode => {
  if (!search) return message;
  const lower = message.toLocaleLowerCase();
  const needle = search.toLocaleLowerCase();
  const parts: ReactNode[] = [];
  let start = 0;
  let match = lower.indexOf(needle);
  while (match >= 0) {
    parts.push(message.slice(start, match));
    parts.push(<mark key={match}>{message.slice(match, match + search.length)}</mark>);
    start = match + search.length;
    match = lower.indexOf(needle, start);
  }
  parts.push(message.slice(start));
  return parts;
};

/** Formats journal microseconds while retaining malformed values as readable text. */
const formatJournalTimestamp = (value: string): string => {
  const timestamp = Number(value);
  return Number.isFinite(timestamp) && timestamp > 0
    ? new Date(timestamp / 1000).toLocaleString()
    : value;
};

/** Displays one structured journal entry with a readable message fallback. */
export const JournalEntryRow = ({
  entry,
  search,
}: {
  entry: JournalEntry;
  search: string;
}): ReactElement => {
  const { getLabel } = useLanguage();
  const date = formatJournalTimestamp(entry.timestamp);
  return (
    <Paper
      data-test-class="journal-entry"
      data-test-name={entry.cursor}
      variant="outlined"
      sx={{ p: 1.5 }}
    >
      <Stack spacing={0.5}>
        <Typography variant="caption" color="text.secondary">
          {date} · {entry.unit || entry.identifier || getLabel("admin.journal.unknownSource")}
          {entry.priority ? ` · ${getLabel("admin.journal.priority")}: ${entry.priority}` : ""}
          {entry.pid ? ` · PID ${entry.pid}` : ""}
        </Typography>
        <Typography variant="body2" sx={{ overflowWrap: "anywhere", whiteSpace: "pre-wrap" }}>
          {entry.message
            ? highlight(entry.message, search)
            : getLabel("admin.journal.unavailableMessage")}
        </Typography>
      </Stack>
    </Paper>
  );
};
