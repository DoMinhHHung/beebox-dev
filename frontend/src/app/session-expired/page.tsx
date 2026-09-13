import { AuthShell } from "@/components/auth/AuthShell";
import { AuthCard } from "@/components/auth/AuthCard";
import Link from "next/link";

export const metadata = { title: "Session expired · BeeBox" };

export default function SessionExpiredPage() {
  return (
    <AuthShell>
      <AuthCard>
        <div className="flex flex-col items-center text-center">
          <h1 className="text-2xl font-semibold tracking-tight text-on-surface mb-1.5">Your session has expired</h1>
          <p className="text-sm text-on-surface-variant mb-8">Please sign in again to continue.</p>
          <Link href="/sign-in" className="w-full h-11 bg-primary-container hover:bg-secondary text-on-primary font-medium text-sm rounded-lg flex items-center justify-center transition-colors shadow-sm">
            Sign in
          </Link>
        </div>
      </AuthCard>
    </AuthShell>
  );
}
