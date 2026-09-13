import { redirect } from "next/navigation";
import { AuthShell } from "@/components/auth/AuthShell";
import { AuthCard } from "@/components/auth/AuthCard";
import { getSessionToken, clearSessionCookie } from "@/lib/session";
import { getSession } from "@/lib/beebox";
import { SignOutButton } from "@/components/auth/SignOutButton";

export const metadata = { title: "Welcome · BeeBox" };

export default async function WelcomePage() {
  const token = await getSessionToken();
  if (!token) redirect("/sign-in");
  try {
    await getSession(token);
  } catch {
    await clearSessionCookie();
    redirect("/session-expired");
  }
  return (
    <AuthShell>
      <AuthCard>
        <div className="flex flex-col items-center text-center">
          <h1 className="text-2xl font-semibold tracking-tight text-on-surface mb-1.5">Welcome back</h1>
          <p className="text-sm text-on-surface-variant mb-8">You’re signed in successfully.</p>
          <div className="w-full"><SignOutButton /></div>
        </div>
      </AuthCard>
    </AuthShell>
  );
}
