import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

const SESSION_COOKIE = process.env.SESSION_COOKIE_NAME || "beebox_session";

/**
 * Protect authenticated routes.
 * - /welcome requires a session cookie.
 * Full session validity is still checked server-side in the page (GET /auth/session).
 * This middleware is a fast first gate to avoid flashing protected content.
 */
export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const token = request.cookies.get(SESSION_COOKIE)?.value;

  // Protect /welcome (and any future authenticated routes)
  if (pathname.startsWith("/welcome")) {
    if (!token) {
      const signInUrl = new URL("/sign-in", request.url);
      // Optional: carry the intended destination
      signInUrl.searchParams.set("from", pathname);
      return NextResponse.redirect(signInUrl);
    }
  }

  // If user is already signed in and visits sign-in / sign-up, send them to welcome
  if (token && (pathname === "/sign-in" || pathname === "/sign-up")) {
    return NextResponse.redirect(new URL("/welcome", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/welcome/:path*", "/sign-in", "/sign-up"],
};
