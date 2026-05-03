export interface AuthUser {
  subject: string;
  username: string;
  email: string | null;
  roles: string[];
  [key: string]: unknown;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface LoginResponse {
  accessToken: string;
  tokenType: "Bearer";
  expiresIn: number;
  refreshToken?: string;
  user: AuthUser;
}

export interface RefreshResponse {
  accessToken: string;
  tokenType: "Bearer";
  expiresIn: number;
  refreshToken?: string;
}
