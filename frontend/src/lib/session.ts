import { cookies } from "next/headers";
import { getServerEnv } from "./env";

export async function getSessionToken(): Promise<string | undefined> {
  const cookieStore = await cookies();
  const env = getServerEnv();
  return cookieStore.get(env.SESSION_COOKIE_NAME)?.value;
}

export async function setSessionCookie(token: string, expiresAt?: string) {
  const cookieStore = await cookies();
  const env = getServerEnv();
  cookieStore.set(env.SESSION_COOKIE_NAME, token, {
    httpOnly: true,
    secure: env.SESSION_COOKIE_SECURE ?? env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    expires: expiresAt ? new Date(expiresAt) : undefined,
  });
}

export async function clearSessionCookie() {
  const cookieStore = await cookies();
  const env = getServerEnv();
  cookieStore.delete(env.SESSION_COOKIE_NAME);
}
