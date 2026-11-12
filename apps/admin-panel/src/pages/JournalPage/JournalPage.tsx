import { useApi } from "@hooks/useApi";
import { useRbac } from "@hooks/useRbac";
import { useLanguage } from "@lite-nas/language";
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  MenuItem,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import { JournalEntryRow } from "@pages/JournalPage/JournalEntryRow";
import type { JournalEntry } from "@schemas/management/journal";
import {
  journalQuerySchema,
  journalResponseSchema,
  journalUnitsResponseSchema,
  journalVacuumResponseSchema,
  journalVacuumSchema,
} from "@schemas/management/journal";
import type { ReactElement } from "react";
import { useState } from "react";

/** Converts a browser local date-time input into an absolute UTC filter. */
const utcTime = (value: string): string => (value ? new Date(value).toISOString() : "");

/** Renders the Security Journal page and its chunked journal browser. */
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The page keeps the journal modal and confirmed action in one direct flow.
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: The page keeps the journal modal and confirmed action in one direct flow.
export const JournalPage = (): ReactElement | null => {
  const { requireSecurity } = useRbac();
  const { getLabel } = useLanguage();
  const { get, post } = useApi();
  const [open, setOpen] = useState(false);
  const [vacuumOpen, setVacuumOpen] = useState(false);
  const [entries, setEntries] = useState<JournalEntry[]>([]);
  const [cursor, setCursor] = useState("");
  const [hasMore, setHasMore] = useState(false);
  const [search, setSearch] = useState("");
  const [unit, setUnit] = useState("");
  const [units, setUnits] = useState<string[]>([]);
  const [priority, setPriority] = useState("");
  const [since, setSince] = useState("");
  const [until, setUntil] = useState("");
  const [days, setDays] = useState("30");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [unitError, setUnitError] = useState("");
  const [vacuumError, setVacuumError] = useState("");
  const [vacuumPending, setVacuumPending] = useState(false);

  /** Loads a fresh chunk or appends entries after the current cursor. */
  // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: Query construction and response handling stay together at the API boundary.
  const load = async (more: boolean): Promise<void> => {
    setLoading(true);
    setError("");
    if (!more) setUnitError("");
    try {
      let query = journalQuerySchema.parse({
        since: utcTime(since),
        until: utcTime(until),
        message: search,
        unit,
        priority,
        cursor: more ? cursor : "",
        limit: 50,
      });
      if (query.since && query.until && query.since > query.until) throw new Error("invalid range");
      if (!more) {
        try {
          const unitParams = new URLSearchParams();
          for (const key of ["since", "until", "message", "priority"] as const)
            if (query[key]) unitParams.set(key, query[key]);
          const unitResponse = await get(`/api/journal/units?${unitParams.toString()}`).execute();
          if (!unitResponse.ok) throw new Error("unit query failed");
          const matchingUnits = journalUnitsResponseSchema.parse(await unitResponse.json()).data
            .units;
          setUnits(matchingUnits);
          if (query.unit && !matchingUnits.includes(query.unit)) {
            query = { ...query, unit: "" };
            setUnit("");
          }
        } catch {
          setUnitError(getLabel("admin.journal.unitsLoadFailed"));
        }
      }
      const params = new URLSearchParams();
      for (const [key, value] of Object.entries(query))
        if (value !== "" && key !== "cursor") params.set(key, String(value));
      if (more && cursor) params.set("cursor", cursor);
      const response = await get(`/api/journal?${params.toString()}`).execute();
      if (!response.ok) throw new Error("query failed");
      const data = journalResponseSchema.parse(await response.json()).data;
      setEntries((previous) => (more ? [...previous, ...data.entries] : data.entries));
      setCursor(data.next_cursor ?? "");
      setHasMore(data.has_more);
    } catch {
      setError(getLabel("admin.journal.loadFailed"));
    } finally {
      setLoading(false);
    }
  };

  /** Executes the explicitly confirmed vacuum request. */
  // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: The confirmed request and its pending/error state form one interaction.
  const vacuum = async (): Promise<void> => {
    if (vacuumPending) return;
    setVacuumPending(true);
    setVacuumError("");
    try {
      const payload = journalVacuumSchema.parse({ days: Number(days) });
      const response = await post("/api/journal/vacuum", payload).execute();
      if (!response.ok || !journalVacuumResponseSchema.parse(await response.json()).completed)
        throw new Error("vacuum failed");
      setVacuumOpen(false);
      await load(false);
    } catch {
      setVacuumError(getLabel("admin.journal.vacuumFailed"));
    } finally {
      setVacuumPending(false);
    }
  };

  if (!requireSecurity()) return null;
  return (
    <Stack data-testid="journal-page" spacing={3}>
      <Typography variant="h1">{getLabel("navigation.Journal")}</Typography>
      <Paper sx={{ p: 3 }}>
        <Stack spacing={2}>
          <Typography>{getLabel("admin.journal.summary")}</Typography>
          <Stack direction="row" spacing={2}>
            <Button
              data-testid="journal-browse"
              variant="contained"
              onClick={() => {
                setOpen(true);
                void load(false);
              }}
            >
              {getLabel("admin.journal.browse")}
            </Button>
            <Button
              data-testid="journal-vacuum"
              color="warning"
              onClick={() => setVacuumOpen(true)}
            >
              {getLabel("admin.journal.vacuum")}
            </Button>
          </Stack>
        </Stack>
      </Paper>
      <Dialog
        data-testid="journal-dialog"
        open={open}
        onClose={() => setOpen(false)}
        fullWidth
        maxWidth="lg"
      >
        <DialogTitle>{getLabel("admin.journal.browse")}</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <Stack direction={{ xs: "column", md: "row" }} spacing={1}>
              <TextField
                data-testid="journal-search"
                label={getLabel("admin.journal.search")}
                value={search}
                onChange={(event) => setSearch(event.target.value)}
              />
              <TextField
                data-testid="journal-unit"
                select
                label={getLabel("admin.journal.unit")}
                value={unit}
                onChange={(event) => setUnit(event.target.value)}
                sx={{ minWidth: 180 }}
              >
                <MenuItem value="">{getLabel("admin.journal.anyUnit")}</MenuItem>
                {units.map((name) => (
                  <MenuItem key={name} value={name}>
                    {name}
                  </MenuItem>
                ))}
              </TextField>
              <TextField
                data-testid="journal-priority"
                select
                label={getLabel("admin.journal.priority")}
                value={priority}
                onChange={(event) => setPriority(event.target.value)}
                sx={{ minWidth: 120 }}
              >
                <MenuItem value="">{getLabel("admin.journal.anyPriority")}</MenuItem>
                {[0, 1, 2, 3, 4, 5, 6, 7].map((value) => (
                  <MenuItem key={value} value={String(value)}>
                    {value}
                  </MenuItem>
                ))}
              </TextField>
            </Stack>
            <Stack direction={{ xs: "column", md: "row" }} spacing={1}>
              <TextField
                data-testid="journal-since"
                label={getLabel("admin.journal.since")}
                type="datetime-local"
                value={since}
                onChange={(event) => setSince(event.target.value)}
                slotProps={{ inputLabel: { shrink: true } }}
              />
              <TextField
                data-testid="journal-until"
                label={getLabel("admin.journal.until")}
                type="datetime-local"
                value={until}
                onChange={(event) => setUntil(event.target.value)}
                slotProps={{ inputLabel: { shrink: true } }}
              />
              <Button data-testid="journal-apply-filters" onClick={() => void load(false)}>
                {getLabel("admin.journal.apply")}
              </Button>
            </Stack>
            {error && <Typography role="alert">{error}</Typography>}
            {unitError && <Typography role="alert">{unitError}</Typography>}
            {!loading && !error && entries.length === 0 && (
              <Typography>{getLabel("admin.journal.empty")}</Typography>
            )}
            {entries.map((entry) => (
              <JournalEntryRow key={entry.cursor} entry={entry} search={search} />
            ))}
            {hasMore && (
              <Button
                data-testid="journal-load-more"
                disabled={loading}
                onClick={() => void load(true)}
              >
                {getLabel("admin.journal.loadMore")}
              </Button>
            )}
            {loading && <Typography>{getLabel("admin.journal.loading")}</Typography>}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpen(false)}>{getLabel("admin.journal.close")}</Button>
        </DialogActions>
      </Dialog>
      <Dialog
        data-testid="journal-vacuum-dialog"
        open={vacuumOpen}
        onClose={() => {
          if (!vacuumPending) setVacuumOpen(false);
        }}
      >
        <DialogTitle>{getLabel("admin.journal.vacuum")}</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            <Typography>{getLabel("admin.journal.vacuumWarning")}</Typography>
            <TextField
              data-testid="journal-vacuum-days"
              type="number"
              label={getLabel("admin.journal.days")}
              value={days}
              onChange={(event) => setDays(event.target.value)}
              slotProps={{ htmlInput: { min: 1, max: 365 } }}
            />
            {vacuumError && <Typography role="alert">{vacuumError}</Typography>}
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button disabled={vacuumPending} onClick={() => setVacuumOpen(false)}>
            {getLabel("admin.journal.cancel")}
          </Button>
          <Button
            data-testid="journal-vacuum-confirm"
            disabled={vacuumPending}
            color="warning"
            onClick={() => void vacuum()}
          >
            {getLabel("admin.journal.confirmVacuum")}
          </Button>
        </DialogActions>
      </Dialog>
    </Stack>
  );
};
