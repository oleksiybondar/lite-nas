import { keyLines } from "@components/backup-restore/MergeDiffViewer/line-utils";
import { MergeSourcePane } from "@components/backup-restore/MergeDiffViewer/MergeSourcePane";
import {
  highlightCodeLine,
  syntaxForPath,
} from "@components/backup-restore/MergeDiffViewer/syntax";
import { changedLines } from "@helpers/backup-restore";
import { useLanguage } from "@lite-nas/language";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import ArrowForwardIcon from "@mui/icons-material/ArrowForward";
import Box from "@mui/material/Box";
import Grid from "@mui/material/Grid";
import IconButton from "@mui/material/IconButton";
import Paper from "@mui/material/Paper";
import Stack from "@mui/material/Stack";
import { alpha, useTheme } from "@mui/material/styles";
import Tooltip from "@mui/material/Tooltip";
import Typography from "@mui/material/Typography";
import type { ReactElement, SyntheticEvent } from "react";
import { useRef } from "react";
import { loadLanguages } from "reprism";
import ini from "reprism/languages/ini";
import json from "reprism/languages/json";
import yaml from "reprism/languages/yaml";

loadLanguages(ini, json, yaml);

type MergeDiffViewerProps = {
  /** Snapshot content used as the left merge source. */
  oldValue: string;
  /** Current editable merge result displayed in the center. */
  newValue: string;
  /** Live configuration used as the right merge source. */
  currentValue: string;
  /** Called whenever the center result is edited or receives transferred lines. */
  onChange: (value: string) => void;
  /** Prevents edits while a merge action is being submitted. */
  disabled: boolean;
  /** File path used to select a syntax grammar and label the editor. */
  summary: string;
  /** Localized pane labels. */
  leftTitle: string;
  resultTitle: string;
  rightTitle: string;
};

/** Displays snapshot, editable result, and current configuration in a themed three-way merge. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The three-pane merge keeps source selection and synchronized rendering together.
export const MergeDiffViewer = ({
  oldValue,
  newValue,
  currentValue,
  onChange,
  disabled,
  summary,
  leftTitle,
  resultTitle,
  rightTitle,
}: MergeDiffViewerProps): ReactElement => {
  const { getLabel } = useLanguage();
  const theme = useTheme();
  const resultEditor = useRef<HTMLTextAreaElement>(null);
  const selection = useRef({ start: newValue.length, end: newValue.length });
  const language = syntaxForPath(summary);
  const snapshotChanges = changedLines(oldValue, newValue);
  const currentChanges = changedLines(currentValue, newValue);
  const resultChanges = new Set([
    ...changedLines(oldValue, newValue),
    ...changedLines(currentValue, newValue),
  ]);

  const insertSelection = (source: "snapshot" | "current"): void => {
    const selected = window.getSelection();
    const sourceElement =
      selected?.anchorNode instanceof Element
        ? selected.anchorNode.closest(`[data-merge-source="${source}"]`)
        : selected?.anchorNode?.parentElement?.closest(`[data-merge-source="${source}"]`);
    const selectedText = selected?.toString();
    if (!sourceElement || !selectedText) return;
    const { start, end } = selection.current;
    const merged = `${newValue.slice(0, start)}${selectedText}${newValue.slice(end)}`;
    onChange(merged);
    requestAnimationFrame(() => {
      resultEditor.current?.focus();
      const cursor = start + selectedText.length;
      resultEditor.current?.setSelectionRange(cursor, cursor);
      selection.current = { start: cursor, end: cursor };
    });
  };

  // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The three-pane merge keeps source selection and synchronized rendering together.
  const sourceAction = (source: "snapshot" | "current"): ReactElement => (
    <Tooltip
      title={getLabel(
        `admin.backup.insertSelected${source === "snapshot" ? "Snapshot" : "Current"}Lines`,
      )}
    >
      <span>
        <IconButton
          size="small"
          disabled={disabled}
          aria-label={getLabel(
            `admin.backup.insertSelected${source === "snapshot" ? "Snapshot" : "Current"}Lines`,
          )}
          onMouseDown={(event) => event.preventDefault()}
          onClick={() => insertSelection(source)}
        >
          {source === "snapshot" ? (
            <ArrowForwardIcon fontSize="small" />
          ) : (
            <ArrowBackIcon fontSize="small" />
          )}
        </IconButton>
      </span>
    </Tooltip>
  );

  const resultLines = newValue.split("\n");
  const keyedResultLines = keyLines(resultLines);
  return (
    <Box data-testid="merge-diff-viewer">
      <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
        {summary}
      </Typography>
      <Typography variant="caption" color="text.secondary" sx={{ display: "block", mb: 1 }}>
        {getLabel("admin.backup.selectLinesToMerge")}
      </Typography>
      <Grid container spacing={1}>
        <Grid size={{ xs: 12, md: 4 }}>
          <MergeSourcePane
            title={leftTitle}
            value={oldValue}
            changed={snapshotChanges}
            tone="snapshot"
            language={language}
            testId="merge-snapshot-pane"
            action={sourceAction("snapshot")}
          />
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <Paper
            variant="outlined"
            data-testid="merge-result-pane"
            sx={{ minWidth: 0, overflow: "hidden" }}
          >
            <Stack
              direction="row"
              alignItems="center"
              justifyContent="space-between"
              sx={{ px: 1.5, py: 0.75, borderBottom: 1, borderColor: "divider" }}
            >
              <Typography variant="subtitle2">{resultTitle}</Typography>
            </Stack>
            <Box
              sx={{
                position: "relative",
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
                "& .merge-line.changed": {
                  backgroundColor: alpha(theme.palette.warning.main, 0.14),
                },
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
              <Box
                aria-hidden="true"
                data-testid="merge-result-highlight"
                sx={{ position: "absolute", inset: 0, overflow: "hidden auto", py: 1.5 }}
              >
                {keyedResultLines.map(({ line, index, key }) => (
                  <Box
                    key={key}
                    className={`merge-line${resultChanges.has(index) ? " changed" : ""}`}
                  >
                    <span className="merge-line-number">{index + 1}</span>
                    <span className="merge-line-content">{highlightCodeLine(line, language)}</span>
                  </Box>
                ))}
              </Box>
              <Box
                component="textarea"
                ref={resultEditor}
                data-testid="merge-result"
                aria-label={resultTitle}
                value={newValue}
                disabled={disabled}
                onChange={(event) => onChange(event.target.value)}
                onSelect={(event: SyntheticEvent<HTMLTextAreaElement>) => {
                  selection.current = {
                    start: event.currentTarget.selectionStart,
                    end: event.currentTarget.selectionEnd,
                  };
                }}
                onScroll={(event) => {
                  const highlightPane = event.currentTarget.previousElementSibling;
                  if (highlightPane instanceof HTMLElement) {
                    highlightPane.scrollTop = event.currentTarget.scrollTop;
                    highlightPane.scrollLeft = event.currentTarget.scrollLeft;
                  }
                }}
                spellCheck={false}
                sx={{
                  position: "absolute",
                  inset: 0,
                  width: "100%",
                  height: "100%",
                  resize: "none",
                  overflow: "auto",
                  border: 0,
                  outline: 0,
                  p: 1.5,
                  pl: "4.5em",
                  bgcolor: "transparent",
                  color: "transparent",
                  caretColor: "text.primary",
                  WebkitTextFillColor: "transparent",
                  font: "inherit",
                  lineHeight: "inherit",
                  letterSpacing: "inherit",
                  whiteSpace: "pre",
                  tabSize: 2,
                  "&::selection": { backgroundColor: alpha(theme.palette.primary.main, 0.35) },
                }}
              />
            </Box>
          </Paper>
        </Grid>
        <Grid size={{ xs: 12, md: 4 }}>
          <MergeSourcePane
            title={rightTitle}
            value={currentValue}
            changed={currentChanges}
            tone="current"
            language={language}
            testId="merge-current-pane"
            action={sourceAction("current")}
          />
        </Grid>
      </Grid>
    </Box>
  );
};
