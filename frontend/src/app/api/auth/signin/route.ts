import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { signIn, mapAuthError, BeeboxApiError } from "@/lib/beebox";
import { attachSessionCookie } from "@/lib/session";

const bodySchema = z.object({
  identifier: z.string().min(1),
  password: z.string().min(1),
});

export async function POST(req: NextRequest) {
  let body: unknown;
  try {
    body = await req.json();
  } catch {
    return NextResponse.json(
      { title: "Invalid request", message: "Please try again." },
      { status: 400 }
    );
  }

  const parsed = bodySchema.safeParse(body);
  if (!parsed.success) {
    return NextResponse.json(
      { title: "Invalid input", message: "Please check the form and try again." },
      { status: 400 }
    );
  }

  try {
    const result = await signIn(parsed.data);

    if (!result?.session_id) {
      return NextResponse.json(
        {
          title: "Something went wrong",
          message: "Sign-in succeeded but no session was returned. Please try again.",
        },
        { status: 500 }
      );
    }

    const response = NextResponse.json({ ok: true });
    return attachSessionCookie(response, result.session_id, result.expires_at);
  } catch (err) {
    const mapped = mapAuthError(err);
    const status =
      err instanceof BeeboxApiError
        ? err.status === 401
          ? 401
          : err.status >= 500
            ? 503
            : 400
        : 500;
    return NextResponse.json(
      { title: mapped.title, message: mapped.message },
      { status }
    );
  }
}
