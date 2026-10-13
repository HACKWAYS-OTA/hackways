import { ApiResponse, ApiError } from "@/types/api";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1";

export class ApiClientError extends Error {
  code: string;
  details?: Record<string, unknown>;

  constructor(message: string, code: string, details?: Record<string, unknown>) {
    super(message);
    this.name = "ApiClientError";
    this.code = code;
    this.details = details;
  }
}

export async function fetchApi<T>(path: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_URL}${path.startsWith("/") ? path : `/${path}`}`;

  const headers: HeadersInit = {
    "Content-Type": "application/json",
    Accept: "application/json",
    ...options.headers,
  };

  const response = await fetch(url, {
    ...options,
    headers,
    credentials: "include", // transport session cookies
  });

  if (!response.ok) {
    let errorEnvelope: ApiError | null = null;
    try {
      errorEnvelope = (await response.json()) as ApiError;
    } catch {
      // response was not JSON
    }

    const code = errorEnvelope?.error?.code || `HTTP_${response.status}`;
    const message = errorEnvelope?.error?.message || response.statusText || "Request failed";
    const details = errorEnvelope?.error?.details;

    throw new ApiClientError(message, code, details);
  }

  if (response.status === 204) {
    return {} as T;
  }

  return response.json() as Promise<T>;
}
