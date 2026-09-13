import { NextRequest, NextResponse } from "next/server";
import { clearSessionCookie } from "@/lib/session";

/**
 * Clears the session cookie and redirects to the session-expired page.
 * Used when a protected page detects an invalid/expired session.
 * Cookie mutation is only allowed in Route Handlers / Server Actions.
 */
export async function GET(request: NextRequest) {
  await clearSessionCookie();

  const url = request.nextUrl.clone();
  url.pathname = "/session-expired";
  url.search = "";
  return NextResponse.redirect(url);
}

export async function POST() {
  await clearSessionCookie();
  return NextResponse.json({ ok: true });
}
