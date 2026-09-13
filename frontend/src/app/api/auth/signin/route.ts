import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { signIn, mapAuthError, BeeboxApiError } from "@/lib/beebox";
import { setSessionCookie } from "@/lib/session";

const bodySchema = z.object({ identifier: z.string().min(1), password: z.string().min(1) });

export async function POST(req: NextRequest) {
  let body: unknown;
  try { body = await req.json(); } catch {
    return NextResponse.json({ title: "Invalid request", message: "Please try again." }, { status: 400 });
  }
  const parsed = bodySchema.safeParse(body);
  if (!parsed.success) {
    return NextResponse.json({ title: "Invalid input", message: "Please check the form and try again." }, { status: 400 });
  }
  try {
    const result = await signIn(parsed.data);
    await setSessionCookie(result.session_id, result.expires_at);
    return NextResponse.json({ ok: true });
  } catch (err) {
    const mapped = mapAuthError(err);
    const status = err instanceof BeeboxApiError ? (err.status === 401 ? 401 : err.status >= 500 ? 503 : 400) : 500;
    return NextResponse.json({ title: mapped.title, message: mapped.message }, { status });
  }
}
