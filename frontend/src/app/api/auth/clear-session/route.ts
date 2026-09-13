import { NextRequest, NextResponse } from "next/server";
import { attachClearedSessionCookie } from "@/lib/session";

/**
 * Clears the session cookie and redirects to the session-expired page.
 * Cookie mutation is only allowed in Route Handlers / Server Actions.
 */
export async function GET(request: NextRequest) {
  const url = request.nextUrl.clone();
  url.pathname = "/session-expired";
  url.search = "";

  const response = NextResponse.redirect(url);
  return attachClearedSessionCookie(response);
}

export async function POST() {
  const response = NextResponse.json({ ok: true });
  return attachClearedSessionCookie(response);
}
