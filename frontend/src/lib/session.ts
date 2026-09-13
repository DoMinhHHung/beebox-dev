import { cookies } from "next/headers";
import { NextResponse } from "next/server";
import { getServerEnv } from "./env";

export const DEFAULT_SESSION_COOKIE = "beebox_session";

export async function getSessionToken(): Promise<string | undefined> {
  const cookieStore = await cookies();
  const env = getServerEnv();
  return cookieStore.get(env.SESSION_COOKIE_NAME)?.value;
}

function cookieOptions(expiresAt?: string) {
  const env = getServerEnv();
  const secure = env.SESSION_COOKIE_SECURE ?? env.NODE_ENV === "production";

  let expires: Date | undefined;
  if (expiresAt) {
    const d = new Date(expiresAt);
    if (!Number.isNaN(d.getTime()) && d.getTime() > Date.now()) {
      expires = d;
    }
  }

  return {
    httpOnly: true,
    secure,
    sameSite: "lax" as const,
    path: "/",
    ...(expires ? { expires } : { maxAge: 60 * 60 * 24 * 7 }), // 7 days fallback
  };
}

/** Prefer this in Route Handlers – attaches Set-Cookie on the response. */
export function attachSessionCookie(
  response: NextResponse,
  token: string,
  expiresAt?: string
) {
  const env = getServerEnv();
  response.cookies.set(env.SESSION_COOKIE_NAME, token, cookieOptions(expiresAt));
  return response;
}

/** Prefer this in Route Handlers when clearing. */
export function attachClearedSessionCookie(response: NextResponse) {
  const env = getServerEnv();
  response.cookies.set(env.SESSION_COOKIE_NAME, "", {
    httpOnly: true,
    secure: env.SESSION_COOKIE_SECURE ?? env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: 0,
  });
  return response;
}

/** Still available for Route Handlers that don't build a custom response. */
export async function setSessionCookie(token: string, expiresAt?: string) {
  const cookieStore = await cookies();
  const env = getServerEnv();
  cookieStore.set(env.SESSION_COOKIE_NAME, token, cookieOptions(expiresAt));
}

export async function clearSessionCookie() {
  const cookieStore = await cookies();
  const env = getServerEnv();
  cookieStore.set(env.SESSION_COOKIE_NAME, "", {
    httpOnly: true,
    secure: env.SESSION_COOKIE_SECURE ?? env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: 0,
  });
}
