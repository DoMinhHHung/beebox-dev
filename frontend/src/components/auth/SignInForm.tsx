"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { FormField } from "./FormField";
import { Button } from "./Button";
import { ErrorBanner } from "./ErrorBanner";

export function SignInForm() {
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
      const res = await fetch("/api/auth/signin", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ identifier, password }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        setError({ title: data.title ?? "Incorrect email or password", message: data.message ?? "Check your details and try again." });
        setLoading(false);
        return;
      }
      router.push("/welcome");
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
      <div>
        <FormField id="password" label="Password" type="password" placeholder="••••••••" autoComplete="current-password" required value={password} onChange={(e) => setPassword(e.target.value)} />
        <div className="mt-1.5 text-right">
          <Link href="/forgot-password" className="text-xs font-medium text-primary hover:underline">Forgot password?</Link>
        </div>
      </div>
      <Button type="submit" loading={loading} className="mt-2">Continue</Button>
    </form>
  );
}
