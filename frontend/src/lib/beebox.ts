import { getServerEnv } from "./env";

export class BeeboxApiError extends Error {
  code: string;
  status: number;
  constructor(code: string, message: string, status: number) {
    super(message);
    this.name = "BeeboxApiError";
    this.code = code;
    this.status = status;
  }
}

async function beeboxFetch<T>(
  path: string,
  options: RequestInit = {},
  token?: string
): Promise<T> {
  const env = getServerEnv();
  const base = env.BEEBOX_IDENTITY_URL.replace(/\/$/, "");
  const url = `${base}${path}`;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    Accept: "application/json",
    ...((options.headers as Record<string, string>) || {}),
  };
  if (token) headers["Authorization"] = `Bearer ${token}`;

  if (process.env.NODE_ENV === "development") {
    console.log(`[beebox] → ${options.method || "GET"} ${url}`);
  }

  let response: Response;
  try {
    response = await fetch(url, { ...options, headers, cache: "no-store" });
  } catch (err) {
    console.error(`[beebox] network error calling ${url}:`, err);
    throw new BeeboxApiError(
      "SERVICE_UNAVAILABLE",
      "We couldn’t connect to BeeBox. Please try again.",
      503
    );
  }

  if (response.status === 204) return undefined as T;

  const body = await response.json().catch(() => null);

  if (!response.ok) {
    const code = body?.error?.code ?? "UNKNOWN_ERROR";
    const message =
      body?.error?.message ?? "Something went wrong. Please try again.";
    console.error(
      `[beebox] ← ${response.status} ${url}`,
      body ?? "(no body)"
    );
    throw new BeeboxApiError(code, message, response.status);
  }

  if (process.env.NODE_ENV === "development") {
    console.log(`[beebox] ← ${response.status} ${url}`);
  }

  return body as T;
}

export type SignUpRequest = { identifier: string; password: string };
export type SignInRequest = { identifier: string; password: string };
export type SignInResponse = {
  user_id: string;
  session_id: string;
  expires_at: string;
};
export type SessionResponse = {
  user_id: string;
  session_id: string;
  organization_id: string;
};
export type RequestPasswordResetRequest = { identifier: string };
export type ResetPasswordRequest = {
  reset_id: string;
  token: string;
  new_password: string;
};

export async function signUp(data: SignUpRequest) {
  return beeboxFetch<{ user_id: string }>("/auth/signup", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function signIn(data: SignInRequest) {
  return beeboxFetch<SignInResponse>("/auth/signin", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function getSession(token: string) {
  return beeboxFetch<SessionResponse>("/auth/session", { method: "GET" }, token);
}

export async function signOut(token: string) {
  return beeboxFetch<void>("/auth/signout", { method: "POST" }, token);
}

export async function requestPasswordReset(data: RequestPasswordResetRequest) {
  return beeboxFetch<{ status: string }>("/auth/password-reset/request", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export async function resetPassword(data: ResetPasswordRequest) {
  return beeboxFetch<{ user_id: string }>("/auth/password-reset/reset", {
    method: "POST",
    body: JSON.stringify(data),
  });
}

export function mapAuthError(error: unknown): {
  title: string;
  message: string;
  kind: "credentials" | "validation" | "unavailable" | "generic";
} {
  if (error instanceof BeeboxApiError) {
    if (
      error.status === 401 ||
      error.code === "UNAUTHENTICATED" ||
      error.code === "INVALID_CREDENTIALS"
    ) {
      return {
        title: "Incorrect email or password",
        message: "Check your details and try again.",
        kind: "credentials",
      };
    }
    if (error.status === 409 || error.code === "CONFLICT") {
      return {
        title: "Account already exists",
        message: "Try signing in instead.",
        kind: "validation",
      };
    }
    if (
      error.status === 503 ||
      error.code === "SERVICE_UNAVAILABLE" ||
      error.code === "DEPENDENCY_FAILURE"
    ) {
      return {
        title: "Something went wrong",
        message: "We couldn’t connect to BeeBox. Please try again.",
        kind: "unavailable",
      };
    }
    if (error.status === 400) {
      return {
        title: "Invalid input",
        message: "Please check the form and try again.",
        kind: "validation",
      };
    }
  }
  return {
    title: "Something went wrong",
    message: "Please try again in a moment.",
    kind: "generic",
  };
}
