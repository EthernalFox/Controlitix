import { api } from "@shared/api";
import type {
  AuthUser,
  LoginRequest,
  LoginResponse,
  RefreshResponse
} from "@shared/modules/auth";

export const authApi = {
  login: async (username: string, password: string): Promise<LoginResponse> => {
    const payload: LoginRequest = { username, password };

    return api.post<LoginResponse>("/auth/login", payload);
  },

  refresh: async (): Promise<RefreshResponse> => {
    return api.post<RefreshResponse>("/auth/refresh");
  },

  logout: async (): Promise<void> => {
    await api.post<void>("/auth/logout");
  },

  userinfo: async (): Promise<AuthUser> => {
    return api.get<AuthUser>("/auth/userinfo");
  }
};
