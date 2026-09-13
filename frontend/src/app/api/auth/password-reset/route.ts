import { NextRequest, NextResponse } from "next/server";
import { z } from "zod";
import { requestPasswordReset, resetPassword, mapAuthError, BeeboxApiError } from "@/lib/beebox";

const requestSchema = z.object({ identifier: z.string().min(1) });
const resetSchema = z.object({ reset_id: z.string().min(1), token: z.string().min(1), new_password: z.string().min(8) });

export async function POST(req: NextRequest) {
  let body: unknown;
  try { body = await req.json(); } catch {
    return NextResponse.json({ title: "Invalid request", message: "Please try again." }, { status: 400 });
  }
  if (body && typeof body === "object" && "reset_id" in body) {
    const parsed = resetSchema.safeParse(body);
    if (!parsed.success) {
      return NextResponse.json({ title: "Invalid input", message: "Please check the form and try again." }, { status: 400 });
    }
    try {
      await resetPassword(parsed.data);
      return NextResponse.json({ ok: true });
    } catch (err) {
      const mapped = mapAuthError(err);
      const status = err instanceof BeeboxApiError ? (err.status >= 500 ? 503 : 400) : 500;
      return NextResponse.json({ title: mapped.title, message: mapped.message }, { status });
    }
  }
  const parsed = requestSchema.safeParse(body);
  if (!parsed.success) {
    return NextResponse.json({ title: "Invalid input", message: "Please check the form and try again." }, { status: 400 });
  }
  try {
    await requestPasswordReset(parsed.data);
    return NextResponse.json({ ok: true });
  } catch (err) {
    if (err instanceof BeeboxApiError && err.status >= 500) {
      const mapped = mapAuthError(err);
      return NextResponse.json({ title: mapped.title, message: mapped.message }, { status: 503 });
    }
    return NextResponse.json({ ok: true });
  }
}
