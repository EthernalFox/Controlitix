import { create } from "zustand";

import { decodeJwt } from "@features/auth/lib/decodeJwt";
import {
  ApiRequestError,
  clearAccessToken,
  getAccessToken,
  notifySessionExpired,
  setAccessToken
} from "@shared/api";
import type { AuthUser } from "@shared/modules/auth";

import { authApi } from "./authApi";
import {
  clearTokenRefreshSchedule,
  runRefreshWithMutex,
  scheduleTokenRefresh
} from "./tokenScheduler";

export type AuthStatus =
  | "idle"
  | "authenticating"
  | "authenticated"
  | "unauthenticated";

interface AuthState {
  status: AuthStatus;
  accessToken: string | null;
  expiresAt: number | null;
  user: AuthUser | null;
  lastError: string | null;
  rateLimitUntil: number | null;

  login: (username: string, password: string) => Promise<void>;
  refresh: () => Promise<void>;
  logout: () => Promise<void>;
  bootstrap: () => Promise<void>;
  markUnauthenticated: (errorType?: string | null) => void;
}

const resolveExpiresAt = (accessToken: string, expiresIn: number): number => {
  const payload = decodeJwt(accessToken);
  if (payload?.exp) {
    return payload.exp * 1000;
  }

  return Date.now() + expiresIn * 1000;
};

const mapApiErrorType = (error: unknown): string => {
  if (error instanceof ApiRequestError) {
    return error.payload.type;
  }

  return "/errors/auth/network";
};

const isAuth401Error = (error: unknown): boolean => {
  if (!(error instanceof ApiRequestError)) {
    return false;
  }

  if (error.payload.status === 401) {
    return true;
  }

  return (
    error.payload.type === "/errors/auth/token-expired" ||
    error.payload.type === "/errors/auth/invalid-token"
  );
};

const clearSessionState = (
  set: (
    partial:
      | Partial<AuthState>
      | ((state: AuthState) => Partial<AuthState>)
  ) => void,
  errorType: string | null
): void => {
  clearAccessToken();
  clearTokenRefreshSchedule();

  set({
    status: "unauthenticated",
    accessToken: null,
    expiresAt: null,
    user: null,
    lastError: errorType,
    rateLimitUntil: null
  });
};

export const useAuthStore = create<AuthState>((set, get) => ({
  status: "idle",
  accessToken: null,
  expiresAt: null,
  user: null,
  lastError: null,
  rateLimitUntil: null,

  markUnauthenticated: (errorType) => {
    clearSessionState(set, errorType ?? null);
  },

  login: async (username: string, password: string) => {
    set({
      status: "authenticating",
      lastError: null,
      rateLimitUntil: null
    });

    try {
      const response = await authApi.login(username, password);
      const expiresAt = resolveExpiresAt(response.accessToken, response.expiresIn);

      setAccessToken(response.accessToken);
      scheduleTokenRefresh(expiresAt, async () => {
        try {
          await get().refresh();
        } catch {
          notifySessionExpired();
        }
      });

      set({
        status: "authenticated",
        accessToken: response.accessToken,
        expiresAt,
        user: response.user,
        lastError: null,
        rateLimitUntil: null
      });
    } catch (error) {
      const errorType = mapApiErrorType(error);
      const retryAfter =
        error instanceof ApiRequestError && error.payload.retryAfter
          ? error.payload.retryAfter
          : null;

      clearSessionState(set, errorType);
      set({
        rateLimitUntil: retryAfter ? Date.now() + retryAfter * 1000 : null
      });
      throw error;
    }
  },

  refresh: async () => {
    await runRefreshWithMutex(async () => {
      const previousStatus = get().status;
      if (previousStatus === "idle" || previousStatus === "unauthenticated") {
        set({ status: "authenticating", lastError: null });
      }

      try {
        const response = await authApi.refresh();
        const expiresAt = resolveExpiresAt(response.accessToken, response.expiresIn);

        setAccessToken(response.accessToken);

        const user = await authApi.userinfo();

        scheduleTokenRefresh(expiresAt, async () => {
          try {
            await get().refresh();
          } catch {
            notifySessionExpired();
          }
        });

        set({
          status: "authenticated",
          accessToken: response.accessToken,
          expiresAt,
          user,
          lastError: null,
          rateLimitUntil: null
        });
      } catch (error) {
        clearSessionState(set, mapApiErrorType(error));
        throw error;
      }
    });
  },

  logout: async () => {
    try {
      await authApi.logout();
    } catch {
      // Logout endpoint errors should not block local sign-out.
    }

    clearSessionState(set, null);
  },

  bootstrap: async () => {
    set({
      status: "authenticating",
      lastError: null
    });

    const accessToken = getAccessToken();
    const accessTokenPayload = accessToken ? decodeJwt(accessToken) : null;
    const accessTokenExpiresAt = accessTokenPayload?.exp
      ? accessTokenPayload.exp * 1000
      : null;

    if (accessToken) {
      set({
        accessToken,
        expiresAt: accessTokenExpiresAt
      });
    }

    try {
      const user = await authApi.userinfo();

      scheduleTokenRefresh(accessTokenExpiresAt, async () => {
        try {
          await get().refresh();
        } catch {
          notifySessionExpired();
        }
      });

      set({
        status: "authenticated",
        user,
        lastError: null,
        rateLimitUntil: null
      });

      return;
    } catch (error) {
      if (!isAuth401Error(error)) {
        clearSessionState(set, mapApiErrorType(error));
        return;
      }

      clearAccessToken();

      set({
        accessToken: null,
        expiresAt: null
      });
    }

    try {
      await get().refresh();
    } catch (error) {
      clearSessionState(set, mapApiErrorType(error));
    }
  }
}));
