import type { PropsWithChildren } from "react";
import { Navigate, useLocation } from "react-router";

import { useAuthStore } from "@/features/auth";
import { Group, Loader } from "@/shared/ui/components";

import { APP_PATHS } from "./paths";

const resolveFromPath = (pathname: string, search: string, hash: string): string => {
  return `${pathname}${search}${hash}`;
};

export const ProtectedRoute = ({ children }: PropsWithChildren) => {
  const location = useLocation();
  const status = useAuthStore((state) => state.status);

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
        to={APP_PATHS.LOGIN}
        replace
        state={{ from: resolveFromPath(location.pathname, location.search, location.hash) }}
      />
    );
  }

  return children;
};
