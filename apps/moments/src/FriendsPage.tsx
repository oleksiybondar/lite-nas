import { useLanguage } from "@lite-nas/language";
import AddIcon from "@mui/icons-material/Add";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";
import SearchIcon from "@mui/icons-material/Search";
import Alert from "@mui/material/Alert";
import Avatar from "@mui/material/Avatar";
import Box from "@mui/material/Box";
import Button from "@mui/material/Button";
import Card from "@mui/material/Card";
import CardContent from "@mui/material/CardContent";
import CircularProgress from "@mui/material/CircularProgress";
import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogTitle from "@mui/material/DialogTitle";
import IconButton from "@mui/material/IconButton";
import InputAdornment from "@mui/material/InputAdornment";
import Pagination from "@mui/material/Pagination";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { FormEvent, ReactElement } from "react";
import { useEffect, useState } from "react";
import type { z } from "zod";
import { useFriends } from "./FriendsContext";
import { loadFriendAvatar } from "./moments-data";
import { friendInputSchema, type friendSchema } from "./schemas";

type Friend = z.infer<typeof friendSchema>;
const pageSize = 12;

/** Loads a saved login's public avatar and falls back when it cannot be retrieved. */
const FriendAvatar = ({ username }: { username: string }): ReactElement => {
  const [src, setSrc] = useState<string>();
  useEffect(() => {
    let active = true;
    let objectUrl: string | undefined;
    void loadFriendAvatar(username)
      // biome-ignore lint/complexity/noExcessiveCognitiveComplexity: avatar fallback, URL creation, and unmount cleanup share this promise lifecycle.
      .then((blob) => {
        if (!blob) return;
        const url = URL.createObjectURL(blob);
        if (active) {
          objectUrl = url;
          setSrc(url);
        } else {
          URL.revokeObjectURL(url);
        }
      })
      .catch(() => undefined);
    return () => {
      active = false;
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [username]);
  return <Avatar src={src} alt="" />;
};

/** Shows one database contact and its remove action in the Friends card grid. */
const FriendCard = ({
  friend,
  onRemove,
}: {
  friend: Friend;
  onRemove: () => void;
}): ReactElement => {
  const { getLabel } = useLanguage();
  const name =
    friend.nickname ||
    [friend.firstName, friend.lastName].filter(Boolean).join(" ") ||
    friend.username;
  return (
    <Card variant="outlined" data-test-class="moments-friend-card" data-test-name={friend.username}>
      <CardContent sx={{ display: "flex", alignItems: "center", gap: 1.5 }}>
        <FriendAvatar username={friend.username} />
        <Stack sx={{ minWidth: 0, flex: 1 }}>
          <Typography fontWeight={600} noWrap>
            {name}
          </Typography>
          <Typography color="text.secondary" variant="body2" noWrap>
            {friend.username}
          </Typography>
        </Stack>
        <IconButton
          aria-label={`${getLabel("moments.friends.remove")} ${friend.username}`}
          data-testid={`moments-friend-remove-${friend.username}`}
          onClick={onRemove}
          size="small"
        >
          <DeleteOutlineIcon />
        </IconButton>
      </CardContent>
    </Card>
  );
};

/** Captures a validated contact in a modal before storing it in the shared Friends context. */
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: the modal keeps form validation and submission feedback together.
const AddFriendDialog = ({
  open,
  onClose,
  onSave,
}: {
  open: boolean;
  onClose: () => void;
  onSave: (friend: z.infer<typeof friendInputSchema>) => Promise<void>;
}): ReactElement => {
  const { getLabel } = useLanguage();
  const [username, setUsername] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [nickname, setNickname] = useState("");
  const [busy, setBusy] = useState(false);
  const [failed, setFailed] = useState(false);
  const [invalid, setInvalid] = useState(false);

  const submit = async (event: FormEvent<HTMLFormElement>): Promise<void> => {
    event.preventDefault();
    const parsed = friendInputSchema.safeParse({ username, firstName, lastName, nickname });
    if (!parsed.success) {
      setInvalid(true);
      return;
    }
    setBusy(true);
    setFailed(false);
    setInvalid(false);
    try {
      await onSave(parsed.data);
      setUsername("");
      setFirstName("");
      setLastName("");
      setNickname("");
    } catch {
      setFailed(true);
    } finally {
      setBusy(false);
    }
  };

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="sm">
      <Box component="form" onSubmit={(event) => void submit(event)}>
        <DialogTitle>{getLabel("moments.friends.addTitle")}</DialogTitle>
        <DialogContent>
          <Stack spacing={2} sx={{ pt: 1 }}>
            {failed && <Alert severity="error">{getLabel("moments.friends.saveFailed")}</Alert>}
            {invalid && <Alert severity="error">{getLabel("moments.friends.invalidInput")}</Alert>}
            <TextField
              autoFocus
              required
              label={getLabel("moments.friends.username")}
              value={username}
              onChange={(event) => setUsername(event.target.value)}
            />
            <TextField
              label={getLabel("moments.friends.firstName")}
              value={firstName}
              onChange={(event) => setFirstName(event.target.value)}
            />
            <TextField
              label={getLabel("moments.friends.lastName")}
              value={lastName}
              onChange={(event) => setLastName(event.target.value)}
            />
            <TextField
              label={getLabel("moments.friends.nickname")}
              value={nickname}
              onChange={(event) => setNickname(event.target.value)}
            />
          </Stack>
        </DialogContent>
        <DialogActions>
          <Button onClick={onClose}>{getLabel("moments.friends.cancel")}</Button>
          <Button type="submit" disabled={busy || !username.trim()} variant="contained">
            {getLabel("moments.friends.add")}
          </Button>
        </DialogActions>
      </Box>
    </Dialog>
  );
};

/** Provides searchable, paginated database contacts and an add-contact modal. */
// biome-ignore lint/complexity/noExcessiveCognitiveComplexity: the page coordinates search, pagination, loading, and mutation feedback.
// biome-ignore lint/complexity/noExcessiveLinesPerFunction: the page's related controls and result states form one user workflow.
export const FriendsPage = (): ReactElement => {
  const { getLabel } = useLanguage();
  const { pageData, loading, loadFailed, loadFriends, save, remove } = useFriends();
  const [page, setPage] = useState(1);
  const [searchDraft, setSearchDraft] = useState("");
  const [search, setSearch] = useState("");
  const [addOpen, setAddOpen] = useState(false);
  const [removeFailed, setRemoveFailed] = useState(false);
  const friends = pageData?.items ?? [];
  const metadata = pageData?.metadata;

  useEffect(() => {
    const timer = window.setTimeout(() => {
      const value = searchDraft.trim();
      if (value !== search) {
        setSearch(value);
        setPage(1);
      }
    }, 250);
    return () => window.clearTimeout(timer);
  }, [searchDraft, search]);

  useEffect(() => {
    void loadFriends(page, pageSize, search);
  }, [page, search, loadFriends]);

  const saveFriend = async (friend: z.infer<typeof friendInputSchema>): Promise<void> => {
    await save(friend);
    setAddOpen(false);
  };

  const removeFriend = async (username: string): Promise<void> => {
    setRemoveFailed(false);
    try {
      const nextPage = await remove(username);
      setPage(nextPage);
    } catch {
      setRemoveFailed(true);
    }
  };

  const startResult =
    metadata && metadata.total_count > 0 ? (metadata.page - 1) * metadata.size + 1 : 0;
  const endResult = metadata ? Math.min(metadata.page * metadata.size, metadata.total_count) : 0;

  return (
    <Stack spacing={2.5} data-testid="moments-friends-page" sx={{ width: "100%" }}>
      <Stack direction={{ xs: "column", sm: "row" }} justifyContent="space-between" spacing={2}>
        <Box>
          <Typography variant="h4">{getLabel("moments.friends.title")}</Typography>
          <Typography color="text.secondary" variant="body2" sx={{ mt: 0.5 }}>
            {getLabel("moments.friends.hint")}
          </Typography>
        </Box>
        <Button
          data-testid="moments-add-friend-button"
          onClick={() => setAddOpen(true)}
          startIcon={<AddIcon />}
          variant="contained"
          sx={{ alignSelf: { sm: "flex-start" } }}
        >
          {getLabel("moments.friends.add")}
        </Button>
      </Stack>

      <TextField
        label={getLabel("moments.friends.search")}
        value={searchDraft}
        onChange={(event) => setSearchDraft(event.target.value)}
        inputProps={{ maxLength: 160 }}
        InputProps={{
          startAdornment: (
            <InputAdornment position="start">
              <SearchIcon />
            </InputAdornment>
          ),
        }}
        sx={{ maxWidth: 520 }}
      />

      {removeFailed && <Alert severity="error">{getLabel("moments.friends.removeFailed")}</Alert>}
      {loadFailed && (
        <Alert
          severity="error"
          action={
            <Button onClick={() => void loadFriends(page, pageSize, search)}>
              {getLabel("moments.friends.retry")}
            </Button>
          }
        >
          {getLabel("moments.friends.loadFailed")}
        </Alert>
      )}
      {loading && friends.length === 0 && (
        <Stack alignItems="center" sx={{ py: 5 }}>
          <CircularProgress aria-label={getLabel("moments.friends.loading")} />
        </Stack>
      )}
      {!loading && !loadFailed && friends.length === 0 && (
        <Typography color="text.secondary">
          {search ? getLabel("moments.friends.noResults") : getLabel("moments.friends.empty")}
        </Typography>
      )}
      {friends.length > 0 && (
        <>
          <Stack spacing={1.5} sx={{ width: "100%", maxWidth: 900 }}>
            {friends.map((friend) => (
              <FriendCard
                key={friend.username}
                friend={friend}
                onRemove={() => void removeFriend(friend.username)}
              />
            ))}
          </Stack>
          {metadata && metadata.total_pages > 0 && (
            <Stack
              direction={{ xs: "column", sm: "row" }}
              justifyContent="space-between"
              alignItems="center"
              spacing={1.5}
            >
              <Typography color="text.secondary" variant="body2">
                {getLabel("moments.friends.resultRange")
                  .replace("{start}", String(startResult))
                  .replace("{end}", String(endResult))
                  .replace("{total}", String(metadata.total_count))}
              </Typography>
              <Pagination
                aria-label={getLabel("moments.friends.pagination")}
                count={metadata.total_pages}
                page={page}
                onChange={(_event, value) => setPage(value)}
                color="primary"
              />
            </Stack>
          )}
        </>
      )}
      <AddFriendDialog open={addOpen} onClose={() => setAddOpen(false)} onSave={saveFriend} />
    </Stack>
  );
};
