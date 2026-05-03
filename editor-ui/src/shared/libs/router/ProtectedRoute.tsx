import type { PropsWithChildren } from "react";
import { Navigate, useLocation } from "react-router";

import { hasEditorRole, useAuthStore } from "@features/auth";
import { Group, Loader } from "@shared/ui";

import { routePaths } from "./routes";

const resolveFromPath = (pathname: string, search: string, hash: string): string => {
  return `${pathname}${search}${hash}`;
};

export const ProtectedRoute = ({ children }: PropsWithChildren) => {
  const location = useLocation();
  const status = useAuthStore((state) => state.status);
  const user = useAuthStore((state) => state.user);

  if (status === "idle" || status === "authenticating") {
    return (
      <Group justify="center" align="center" style={{ minHeight: "100vh" }}>
        <Loader />
      </Group>
    );
  }

  if (status !== "authenticated") {
    return (
      <Navigate
        to={routePaths.login}
        replace
        state={{ from: resolveFromPath(location.pathname, location.search, location.hash) }}
      />
    );
  }

  if (!hasEditorRole(user)) {
    return <Navigate to={routePaths.accessDenied} replace />;
  }

  return children;
};
