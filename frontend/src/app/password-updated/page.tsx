import { AuthShell } from "@/components/auth/AuthShell";
import { AuthCard } from "@/components/auth/AuthCard";
import Link from "next/link";

export const metadata = { title: "Password updated · BeeBox" };

export default function PasswordUpdatedPage() {
  return (
    <AuthShell>
      <AuthCard>
        <div className="flex flex-col items-center text-center">
          <div className="mb-6 flex items-center justify-center">
            <div className="w-12 h-12 rounded-full bg-surface-container-low flex items-center justify-center text-primary shadow-sm">
              <svg className="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
          </div>
          <h1 className="text-2xl font-semibold tracking-tight text-on-surface mb-1.5">Password updated</h1>
          <p className="text-sm text-on-surface-variant mb-8">Your password has been changed successfully.</p>
          <Link href="/sign-in" className="w-full h-11 bg-primary-container hover:bg-secondary text-on-primary font-medium text-sm rounded-lg flex items-center justify-center transition-colors shadow-sm">
            Continue to sign in
          </Link>
        </div>
      </AuthCard>
    </AuthShell>
  );
}
