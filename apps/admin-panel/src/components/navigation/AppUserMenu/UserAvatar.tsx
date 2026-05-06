import type { AuthMeUserDTO } from "@dto/auth/auth";
import Avatar from "@mui/material/Avatar";
import type { ReactElement } from "react";
import { resolveUserDisplayName, resolveUserInitials } from "./helpers";

/**
 * Avatar shown for the current authenticated user.
 */
export const UserAvatar = ({ user }: { user: AuthMeUserDTO }): ReactElement => {
  return (
    <Avatar alt={resolveUserDisplayName(user)} src={user.avatar_url} sx={{ height: 34, width: 34 }}>
      {resolveUserInitials(user)}
    </Avatar>
  );
};
