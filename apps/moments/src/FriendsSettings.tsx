import { useLanguage } from "@lite-nas/language";
import DeleteOutlineIcon from "@mui/icons-material/DeleteOutline";
import Avatar from "@mui/material/Avatar";
import Button from "@mui/material/Button";
import IconButton from "@mui/material/IconButton";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { ReactElement } from "react";
import { useCallback, useEffect, useState } from "react";
import type { z } from "zod";
import { listFriends, loadFriendAvatar, removeFriend, saveFriend } from "./moments-data";
import type { friendSchema } from "./schemas";

/** Validated contact row rendered by the Moments friends settings screen. */
type Friend = z.infer<typeof friendSchema>;

/** Creates a display URL for an eligible avatar, or no URL for any unavailable response. */
const friendAvatarUrl = async (username: string): Promise<string | undefined> => {
  const blob = await loadFriendAvatar(username);
  return blob ? URL.createObjectURL(blob) : undefined;
};

/** Stores a loaded object URL or revokes it when its avatar row has already unmounted. */
const keepAvatarUrl = (
  active: boolean,
  next: string | undefined,
  setSrc: (url: string) => void,
): void => {
  if (!next) return;
  if (active) setSrc(next);
  else URL.revokeObjectURL(next);
};

/** Loads a saved login's public avatar and treats every unavailable response as the same fallback. */
const FriendAvatar = ({ username }: { username: string }): ReactElement => {
  const [src, setSrc] = useState<string | undefined>();
  useEffect(() => {
    let active = true;
    let objectUrl: string | undefined;
    void friendAvatarUrl(username)
      .then((next) => {
        keepAvatarUrl(active, next, (url) => {
          objectUrl = next;
          setSrc(url);
        });
      })
      .catch(() => undefined);
    return () => {
      active = false;
      if (objectUrl) URL.revokeObjectURL(objectUrl);
    };
  }, [username]);
  return <Avatar src={src} alt="" />;
};

/** Renders the saved contact row with its optional public avatar and remove action. */
const FriendRow = ({
  friend,
  onRemove,
}: {
  friend: Friend;
  onRemove: (username: string) => void;
}): ReactElement => {
  const { getLabel } = useLanguage();
  const displayName =
    friend.nickname ||
    [friend.firstName, friend.lastName].filter(Boolean).join(" ") ||
    friend.username;
  return (
    <Stack direction="row" spacing={1} alignItems="center">
      <FriendAvatar username={friend.username} />
      <Stack flex={1}>
        <Typography>{displayName}</Typography>
        <Typography color="text.secondary" variant="body2">
          {friend.username}
        </Typography>
      </Stack>
      <IconButton
        aria-label={getLabel("moments.friends.remove")}
        onClick={() => onRemove(friend.username)}
      >
        <DeleteOutlineIcon />
      </IconButton>
    </Stack>
  );
};

/** Captures user-maintained account labels and persists them without account lookup. */
const FriendForm = ({
  onSaved,
  onError,
}: {
  onSaved: () => Promise<void>;
  onError: () => void;
}): ReactElement => {
  const { getLabel } = useLanguage();
  const [username, setUsername] = useState("");
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [nickname, setNickname] = useState("");
  const [busy, setBusy] = useState(false);
  const nameSetters = { firstName: setFirstName, lastName: setLastName, nickname: setNickname };
  const save = async (): Promise<void> => {
    setBusy(true);
    try {
      await saveFriend({ username, firstName, lastName, nickname });
      setUsername("");
      setFirstName("");
      setLastName("");
      setNickname("");
      await onSaved();
    } catch {
      onError();
    } finally {
      setBusy(false);
    }
  };
  return (
    <Stack spacing={2}>
      <TextField
        label={getLabel("moments.friends.username")}
        value={username}
        onChange={(event) => setUsername(event.target.value)}
      />
      <FriendNameFields
        firstName={firstName}
        lastName={lastName}
        nickname={nickname}
        onChange={(field, value) => nameSetters[field](value)}
      />
      <Button disabled={busy || !username.trim()} onClick={() => void save()}>
        {getLabel("moments.friends.save")}
      </Button>
    </Stack>
  );
};

/** Shows the three optional display-name fields without changing the stored login. */
const FriendNameFields = ({
  firstName,
  lastName,
  nickname,
  onChange,
}: {
  firstName: string;
  lastName: string;
  nickname: string;
  onChange: (field: "firstName" | "lastName" | "nickname", value: string) => void;
}): ReactElement => {
  const { getLabel } = useLanguage();
  return (
    <Stack direction={{ xs: "column", sm: "row" }} spacing={2}>
      <TextField
        label={getLabel("moments.friends.firstName")}
        value={firstName}
        onChange={(event) => onChange("firstName", event.target.value)}
      />
      <TextField
        label={getLabel("moments.friends.lastName")}
        value={lastName}
        onChange={(event) => onChange("lastName", event.target.value)}
      />
      <TextField
        label={getLabel("moments.friends.nickname")}
        value={nickname}
        onChange={(event) => onChange("nickname", event.target.value)}
      />
    </Stack>
  );
};

/** Lets users maintain private contact labels and inspect avatars when File Service permits it. */
export const FriendsSettings = (): ReactElement => {
  const { getLabel } = useLanguage();
  const [friends, setFriends] = useState<Friend[]>([]);
  const [error, setError] = useState(false);

  const reload = useCallback(async (): Promise<void> => {
    try {
      setFriends(await listFriends());
      setError(false);
    } catch {
      setError(true);
    }
  }, []);
  useEffect(() => {
    void reload();
  }, [reload]);

  const remove = async (login: string): Promise<void> => {
    try {
      await removeFriend(login);
      setFriends((current) => current.filter((friend) => friend.username !== login));
    } catch {
      setError(true);
    }
  };

  return (
    <Stack spacing={2}>
      <Typography variant="h5">{getLabel("moments.settings.friends")}</Typography>
      <Typography color="text.secondary" variant="body2">
        {getLabel("moments.settings.friendsHint")}
      </Typography>
      <FriendForm onSaved={reload} onError={() => setError(true)} />
      {error && <Typography color="error">{getLabel("moments.friends.failed")}</Typography>}
      {friends.length === 0 && !error && (
        <Typography color="text.secondary">{getLabel("moments.friends.empty")}</Typography>
      )}
      {friends.map((friend) => (
        <FriendRow key={friend.username} friend={friend} onRemove={(login) => void remove(login)} />
      ))}
    </Stack>
  );
};
