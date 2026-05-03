import type { PropsWithChildren } from "react";
import { useEffect } from "react";

import { useAuthStore } from "@features/auth";
import {
  setRefreshAccessTokenHandler,
  setSessionExpiredHandler,
  setUnauthenticatedHandler
} from "@shared/api";
import { routePaths } from "@shared/libs/router";
import { router } from "@shared/libs/router/router";
import { Loader } from "@shared/ui/components";

export const AuthProvider = ({ children }: PropsWithChildren) => {
  const status = useAuthStore((state) => state.status);

  useEffect(() => {
    setRefreshAccessTokenHandler(async () => {
      await useAuthStore.getState().refresh();
    });

    setUnauthenticatedHandler(() => {
      useAuthStore.getState().markUnauthenticated(null);
      void router.navigate(routePaths.login, { replace: true });
    });

    setSessionExpiredHandler(() => {
      useAuthStore.getState().markUnauthenticated(null);
      void router.navigate(routePaths.sessionExpired, { replace: true });
    });

    if (useAuthStore.getState().status === "idle") {
      void useAuthStore.getState().bootstrap();
    }

    return () => {
      setRefreshAccessTokenHandler(null);
      setUnauthenticatedHandler(null);
      setSessionExpiredHandler(null);
    };
  }, []);

  if (status === "idle" || status === "authenticating") {
    return (
      <div
        style={{
          display: "flex",
          alignItems: "center",
          justifyContent: "center",
          minHeight: "100vh"
        }}
      >
        <Loader size="xl" />
      </div>
    );
  }

  return children;
};
