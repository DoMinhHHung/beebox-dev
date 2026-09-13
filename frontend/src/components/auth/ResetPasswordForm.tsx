"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { FormField } from "./FormField";
import { Button } from "./Button";
import { ErrorBanner } from "./ErrorBanner";

type Props = { resetId: string; token: string };

export function ResetPasswordForm({ resetId, token }: Props) {
  const router = useRouter();
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState<{ title: string; message: string } | null>(null);
  const [loading, setLoading] = useState(false);
  const missingParams = !resetId || !token;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    if (password !== confirm) {
      setError({ title: "Passwords don’t match", message: "Please make sure both fields are the same." });
      return;
    }
    if (password.length < 8) {
      setError({ title: "Password too short", message: "Use at least 8 characters." });
      return;
    }
    setLoading(true);
    try {
      const res = await fetch("/api/auth/password-reset", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ reset_id: resetId, token, new_password: password }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        setError({ title: data.title ?? "Could not reset password", message: data.message ?? "Please try again or request a new link." });
        setLoading(false);
        return;
      }
      router.push("/password-updated");
    } catch {
      setError({ title: "Something went wrong", message: "We couldn’t connect. Please try again." });
      setLoading(false);
    }
  }

  if (missingParams) {
    return (
      <div className="text-center space-y-4">
        <ErrorBanner title="Invalid or expired link" message="Please request a new password reset link." />
        <Button variant="secondary" onClick={() => router.push("/forgot-password")}>Request new link</Button>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4" noValidate>
      {error && <ErrorBanner title={error.title} message={error.message} />}
      <FormField id="new-password" label="New password" type="password" placeholder="At least 8 characters" autoComplete="new-password" required value={password} onChange={(e) => setPassword(e.target.value)} />
      <FormField id="confirm-password" label="Confirm password" type="password" placeholder="Repeat your password" autoComplete="new-password" required value={confirm} onChange={(e) => setConfirm(e.target.value)} />
      <p className="text-xs text-on-surface-variant -mt-1">Use at least 8 characters.</p>
      <Button type="submit" loading={loading} className="mt-2">Reset password</Button>
    </form>
  );
}
