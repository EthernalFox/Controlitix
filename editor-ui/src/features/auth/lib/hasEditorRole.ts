import type { AuthUser } from "@shared/modules/auth";

export const hasEditorRole = (user: AuthUser | null): boolean => {
  if (!user) {
    return false;
  }

  return user.roles.some((role) => role === "admin" || role === "engineer");
};
