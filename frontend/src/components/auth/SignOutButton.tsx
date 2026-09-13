"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { Button } from "./Button";

export function SignOutButton() {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  async function handleSignOut() {
    setLoading(true);
    try {
      await fetch("/api/auth/signout", { method: "POST" });
    } finally {
      router.push("/sign-in");
      router.refresh();
    }
  }
  return <Button variant="secondary" loading={loading} onClick={handleSignOut}>Sign out</Button>;
}
