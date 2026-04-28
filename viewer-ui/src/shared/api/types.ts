export interface ApiFieldError {
  field: string;
  message: string;
}

export interface ApiError {
  type: string;
  title: string;
  status: number;
  detail?: string;
  instance?: string;
  errors?: ApiFieldError[];
  retryAfter?: number;
}

export class ApiRequestError extends Error {
  public readonly payload: ApiError;

  public constructor(payload: ApiError) {
    super(payload.detail || payload.title);
    this.name = "ApiRequestError";
    this.payload = payload;
  }
}
