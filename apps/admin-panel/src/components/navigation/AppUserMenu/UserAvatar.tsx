import type { AuthMeUserDTO } from "@dto/auth/auth";
import { useProfileAvatar } from "@hooks/useProfileAvatar";
import Avatar from "@mui/material/Avatar";
import type { ReactElement } from "react";
import { resolveUserDisplayName, resolveUserInitials } from "./helpers";

/**
 * Avatar shown for the current authenticated user.
 */
export const UserAvatar = ({ user }: { user: AuthMeUserDTO }): ReactElement => {
  const profileImage = useProfileAvatar();
  return (
    <Avatar
      alt={resolveUserDisplayName(user)}
      data-testid="user-avatar"
      src={profileImage ?? user.avatar_url}
      sx={{ height: 34, width: 34 }}
    >
      {resolveUserInitials(user)}
    </Avatar>
  );
};
