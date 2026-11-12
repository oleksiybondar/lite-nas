import { keyLines } from "@components/backup-restore/MergeDiffViewer/line-utils";
import { highlightCodeLine } from "@components/backup-restore/MergeDiffViewer/syntax";
import Box from "@mui/material/Box";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import { alpha, useTheme } from "@mui/material/styles";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";

type MergeSourcePaneProps = {
  /** Pane title and transfer action accessible name. */
  title: string;
  /** Original snapshot or current configuration text. */
  value: string;
  /** Zero-based source line indexes that differ from the merge result. */
  changed: Set<number>;
  /** Semantic tint distinguishes snapshot and current changes. */
  tone: "snapshot" | "current";
  /** Syntax grammar selected for the file. */
  language: string | null;
  /** Stable test selector for this source pane. */
  testId: string;
  /** Arrow action that transfers selected text to the result pane. */
  action: ReactElement;
};

/** Displays a selectable, syntax-highlighted source pane with dynamic difference coloring. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The pane keeps line highlighting and source selection in one render flow.
export const MergeSourcePane = ({
  title,
  value,
  changed,
  tone,
  language,
  testId,
  action,
}: MergeSourcePaneProps): ReactElement => {
  const theme = useTheme();
  const color = tone === "snapshot" ? theme.palette.error.main : theme.palette.success.main;

  return (
    <Paper variant="outlined" data-testid={testId} sx={{ minWidth: 0, overflow: "hidden" }}>
      <Stack
        direction="row"
        alignItems="center"
        justifyContent="space-between"
        sx={{ px: 1.5, py: 0.75, borderBottom: 1, borderColor: "divider" }}
      >
        <Typography variant="subtitle2" noWrap title={title}>
          {title}
        </Typography>
        {action}
      </Stack>
      <Box
        data-merge-source={tone}
        sx={{
          overflow: "auto",
          height: { xs: 280, md: "55vh" },
          bgcolor: "background.paper",
          color: "text.primary",
          fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
          fontSize: 12,
          lineHeight: 1.6,
          "& .merge-line": { display: "flex", minHeight: "1.6em", whiteSpace: "pre" },
          "& .merge-line-number": {
            flex: "0 0 3.5em",
            pr: 1,
            color: "text.disabled",
            textAlign: "right",
            userSelect: "none",
          },
          "& .merge-line-content": { flex: 1, pr: 2 },
          "& .merge-line.changed": { backgroundColor: alpha(color, 0.14) },
          "& .token.comment": { color: "text.secondary", fontStyle: "italic" },
          "& .token.property, & .token.selector, & .token.atrule": {
            color: theme.palette.primary.main,
          },
          "& .token.string, & .token.attr-value": { color: theme.palette.success.main },
          "& .token.number, & .token.boolean, & .token.constant": {
            color: theme.palette.warning.main,
          },
          "& .token.punctuation, & .token.operator": { color: "text.secondary" },
          "& .token.important": { color: theme.palette.error.main },
        }}
      >
        {keyLines(value.split("\n")).map(({ line, index, key }) => (
          <Box
            key={key}
            className={`merge-line${changed.has(index) ? " changed" : ""}`}
            data-merge-line={index}
          >
            <span className="merge-line-number">{index + 1}</span>
            <span className="merge-line-content">{highlightCodeLine(line, language)}</span>
          </Box>
        ))}
      </Box>
    </Paper>
  );
};
