export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  offset: number;
  limit: number;
}

export interface ApiFieldError {
  field: string;
  message: string;
}

export interface ApiError {
  type: string;
  title: string;
  status: number;
  detail: string;
  errors?: ApiFieldError[];
}

export class ApiRequestError extends Error {
  public readonly payload: ApiError;

  public constructor(payload: ApiError) {
    super(payload.detail || payload.title);
    this.name = "ApiRequestError";
    this.payload = payload;
  }
}
