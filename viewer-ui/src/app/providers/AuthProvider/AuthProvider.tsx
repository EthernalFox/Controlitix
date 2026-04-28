import type { PropsWithChildren } from "react";
import { useEffect } from "react";

import { router } from "@/app/providers/RouterProvider";
import { useAuthStore } from "@/features/auth";
import {
  setRefreshAccessTokenHandler,
  setSessionExpiredHandler,
  setUnauthenticatedHandler
} from "@/shared/api";
import { APP_PATHS } from "@/shared/libs/router";

export const AuthProvider = ({ children }: PropsWithChildren) => {
  useEffect(() => {
    setRefreshAccessTokenHandler(async () => {
      await useAuthStore.getState().refresh();
    });

    setUnauthenticatedHandler(() => {
      useAuthStore.getState().markUnauthenticated(null);
      void router.navigate(APP_PATHS.LOGIN, { replace: true });
    });

    setSessionExpiredHandler(() => {
      useAuthStore.getState().markUnauthenticated(null);
      void router.navigate(APP_PATHS.SESSION_EXPIRED, { replace: true });
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

  return children;
};
