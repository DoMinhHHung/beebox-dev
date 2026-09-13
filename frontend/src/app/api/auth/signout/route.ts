import { NextResponse } from "next/server";
import { getSessionToken, clearSessionCookie } from "@/lib/session";
import { signOut } from "@/lib/beebox";

export async function POST() {
  const token = await getSessionToken();
  if (token) {
    try { await signOut(token); } catch { /* best-effort */ }
  }
  await clearSessionCookie();
  return NextResponse.json({ ok: true });
}
