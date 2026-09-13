"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { FormField } from "./FormField";
import { Button } from "./Button";
import { ErrorBanner } from "./ErrorBanner";

export function SignUpForm() {
  const router = useRouter();
  const [identifier, setIdentifier] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<{ title: string; message: string } | null>(null);
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError(null);
    setLoading(true);
    try {
      const res = await fetch("/api/auth/signup", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ identifier, password }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        setError({ title: data.title ?? "Could not create account", message: data.message ?? "Please try again." });
        setLoading(false);
        return;
      }
      router.push("/sign-in?created=1");
      router.refresh();
    } catch {
      setError({ title: "Something went wrong", message: "We couldn’t connect. Please try again." });
      setLoading(false);
    }
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-4" noValidate>
      {error && <ErrorBanner title={error.title} message={error.message} />}
      <FormField id="identifier" label="Email" type="email" placeholder="you@example.com" autoComplete="email" required value={identifier} onChange={(e) => setIdentifier(e.target.value)} />
      <FormField id="password" label="Password" type="password" placeholder="At least 8 characters" autoComplete="new-password" required value={password} onChange={(e) => setPassword(e.target.value)} />
      <p className="text-xs text-on-surface-variant -mt-1">Use at least 8 characters.</p>
      <Button type="submit" loading={loading} className="mt-2">Create account</Button>
    </form>
  );
}
