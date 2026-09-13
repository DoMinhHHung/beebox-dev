"use client";

import { useState } from "react";
import { FormField } from "./FormField";
import { Button } from "./Button";
import { ErrorBanner } from "./ErrorBanner";

export function ForgotPasswordForm() {
  const [identifier, setIdentifier] = useState("");
  const [error, setError] = useState<{ title: string; message: string } | null>(null);
  const [success, setSuccess] = useState(false);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const res = await fetch("/api/auth/password-reset", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ identifier }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        setError({ title: data.title ?? "Something went wrong", message: data.message ?? "Please try again." });
        setLoading(false);
        return;
      }
      setSuccess(true);
      setLoading(false);
    } catch {
      setError({ title: "Something went wrong", message: "We couldn’t connect. Please try again." });
      setLoading(false);
    }
  }

  if (success) {
    return (
      <div className="text-center space-y-3">
        <p className="text-sm text-on-surface">If an account exists for that email, you’ll receive a reset link shortly.</p>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4" noValidate>
      {error && <ErrorBanner title={error.title} message={error.message} />}
      <FormField id="identifier" label="Email" type="email" placeholder="you@example.com" autoComplete="email" required value={identifier} onChange={(e) => setIdentifier(e.target.value)} />
      <Button type="submit" loading={loading} className="mt-2">Send reset link</Button>
    </form>
  );
}
