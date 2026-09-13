import { NextResponse } from "next/server";
import { getSessionToken, attachClearedSessionCookie } from "@/lib/session";
import { signOut } from "@/lib/beebox";

export async function POST() {
  const token = await getSessionToken();
  if (token) {
    try {
      await signOut(token);
    } catch {
      // best-effort revoke
    }
  }

  const response = NextResponse.json({ ok: true });
  return attachClearedSessionCookie(response);
}
