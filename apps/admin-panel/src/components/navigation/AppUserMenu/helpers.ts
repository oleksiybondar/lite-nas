import type { AuthMeUserDTO } from "@dto/auth/auth";

/**
 * Resolves the best available display name for avatar accessibility.
 */
export const resolveUserDisplayName = (user: AuthMeUserDTO): string => {
  return user.full_name ?? user.login;
};

/**
 * Builds fallback initials from the user's full name or login.
 */
export const resolveUserInitials = (user: AuthMeUserDTO): string => {
  const source = resolveUserDisplayName(user).trim();

  if (source.length === 0) {
    return "?";
  }

  const initials = source
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? "")
    .join("");

  return initials || source[0]?.toUpperCase() || "?";
};
