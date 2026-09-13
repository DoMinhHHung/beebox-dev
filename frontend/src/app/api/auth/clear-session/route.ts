import { NextResponse } from "next/server";
import { clearSessionCookie } from "@/lib/session";

/**
 * Clears the session cookie and redirects to the session-expired page.
 * Used when a protected page detects an invalid/expired session.
 * Cookie mutation is only allowed in Route Handlers / Server Actions.
 */
export async function GET() {
  await clearSessionCookie();

  return NextResponse.redirect(
    new URL("/session-expired", process.env.NEXT_PUBLIC_APP_URL || "http://localhost:3000")
  );
}

export async function POST() {
  await clearSessionCookie();
  return NextResponse.json({ ok: true });
}
