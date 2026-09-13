import { AuthShell } from "@/components/auth/AuthShell";
import { AuthCard } from "@/components/auth/AuthCard";
import Link from "next/link";

export const metadata = { title: "Service unavailable · BeeBox" };

export default function ServiceUnavailablePage() {
  return (
    <AuthShell>
      <AuthCard>
        <div className="flex flex-col items-center text-center">
          <h1 className="text-2xl font-semibold tracking-tight text-on-surface mb-1.5">Something went wrong</h1>
          <p className="text-sm text-on-surface-variant mb-8">We couldn’t connect to BeeBox. Please try again.</p>
          <Link href="/sign-in" className="w-full h-11 bg-primary-container hover:bg-secondary text-on-primary font-medium text-sm rounded-lg flex items-center justify-center transition-colors shadow-sm">
            Try again
          </Link>
        </div>
      </AuthCard>
    </AuthShell>
  );
}
